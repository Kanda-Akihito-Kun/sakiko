package downloadtargets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

const (
	defaultEndpoint     = "https://www.speedtest.net/api/js/servers?engine=js&https_functional=true&limit=100"
	defaultFetchTimeout = 10 * time.Second
	defaultCacheTTL     = 10 * time.Minute
	defaultUserAgent    = "sakiko/0.1"
)

type Config struct {
	Client       *http.Client
	Endpoint     string
	FetchTimeout time.Duration
	CacheTTL     time.Duration
	Now          func() time.Time
}

type Manager struct {
	client       *http.Client
	endpoint     string
	fetchTimeout time.Duration
	cacheTTL     time.Duration
	now          func() time.Time

	mu        sync.RWMutex
	cached    []interfaces.DownloadTarget
	expiresAt time.Time
}

type speedtestServer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	CC      string `json:"cc"`
	Sponsor string `json:"sponsor"`
	Host    string `json:"host"`
	URL     string `json:"url"`
}

func NewManager(cfg Config) *Manager {
	client := cfg.Client
	if client == nil {
		client = &http.Client{}
	}
	endpoint := strings.TrimSpace(cfg.Endpoint)
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	fetchTimeout := cfg.FetchTimeout
	if fetchTimeout <= 0 {
		fetchTimeout = defaultFetchTimeout
	}
	cacheTTL := cfg.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = defaultCacheTTL
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}

	return &Manager{
		client:       client,
		endpoint:     endpoint,
		fetchTimeout: fetchTimeout,
		cacheTTL:     cacheTTL,
		now:          now,
	}
}

func (m *Manager) List() ([]interfaces.DownloadTarget, error) {
	if targets := m.cachedTargets(); len(targets) > 0 {
		return targets, nil
	}
	return m.Refresh()
}

func (m *Manager) Refresh() ([]interfaces.DownloadTarget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("speedtest target request failed: %s", resp.Status)
	}

	var rawServers []speedtestServer
	if err := json.NewDecoder(resp.Body).Decode(&rawServers); err != nil {
		return nil, fmt.Errorf("decode speedtest targets: %w", err)
	}

	targets := make([]interfaces.DownloadTarget, 0, len(rawServers)+1)
	targets = append(targets, interfaces.DefaultDownloadTarget())

	seen := map[string]struct{}{
		targets[0].ID: {},
	}
	for _, raw := range rawServers {
		target, ok := normalizeTarget(raw)
		if !ok {
			continue
		}
		if _, exists := seen[target.ID]; exists {
			continue
		}
		seen[target.ID] = struct{}{}
		targets = append(targets, target)
	}

	m.mu.Lock()
	m.cached = cloneTargets(targets)
	m.expiresAt = m.now().Add(m.cacheTTL)
	m.mu.Unlock()

	return cloneTargets(targets), nil
}

func (m *Manager) cachedTargets() []interfaces.DownloadTarget {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.cached) == 0 {
		return nil
	}
	if !m.expiresAt.IsZero() && !m.now().Before(m.expiresAt) {
		return nil
	}
	return cloneTargets(m.cached)
}

func normalizeTarget(raw speedtestServer) (interfaces.DownloadTarget, bool) {
	serverID := strings.TrimSpace(raw.ID)
	if serverID == "" {
		return interfaces.DownloadTarget{}, false
	}

	endpoint := strings.TrimSpace(raw.URL)
	if endpoint == "" {
		return interfaces.DownloadTarget{}, false
	}

	downloadURL, host, err := speedtestDownloadURL(endpoint)
	if err != nil {
		return interfaces.DownloadTarget{}, false
	}

	city := strings.TrimSpace(raw.Name)
	country := strings.TrimSpace(raw.Country)

	return interfaces.DownloadTarget{
		ID:          "speedtest-" + serverID,
		Source:      interfaces.DownloadTargetSourceSpeedtest,
		Name:        firstNonEmpty(city, country, "Speedtest Server"),
		City:        city,
		Country:     country,
		CountryCode: strings.TrimSpace(raw.CC),
		Sponsor:     strings.TrimSpace(raw.Sponsor),
		Host:        firstNonEmpty(strings.TrimSpace(raw.Host), host),
		Endpoint:    endpoint,
		DownloadURL: downloadURL,
	}, true
}

func speedtestDownloadURL(endpoint string) (string, string, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return "", "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", "", fmt.Errorf("invalid speedtest endpoint: %q", endpoint)
	}

	path := parsed.EscapedPath()
	if path == "" || path == "/" {
		parsed.Path = "/speedtest/random4000x4000.jpg"
	} else {
		lastSlash := strings.LastIndex(path, "/")
		base := "/"
		if lastSlash >= 0 {
			base = path[:lastSlash+1]
		}
		parsed.Path = base + "random4000x4000.jpg"
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), parsed.Host, nil
}

func cloneTargets(targets []interfaces.DownloadTarget) []interfaces.DownloadTarget {
	if len(targets) == 0 {
		return nil
	}
	cloned := make([]interfaces.DownloadTarget, len(targets))
	copy(cloned, targets)
	return cloned
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
