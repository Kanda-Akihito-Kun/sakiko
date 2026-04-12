package backendprobe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/netx"

	"golang.org/x/net/html/charset"
)

const (
	defaultTTL           = 10 * time.Minute
	defaultProbeTimeout  = 4 * time.Second
	currentIPLookupURL   = "https://ipwho.is/"
	locationLookupFormat = "https://ip.cn/ip/%s.html"
)

var locationPatterns = []*regexp.Regexp{
	regexp.MustCompile(`所在地理位置[\s\S]*?<td[^>]*>\s*([^<]+?)\s*</td>`),
	regexp.MustCompile(`鎵€鍦ㄥ湴鐞嗕綅缃[\s\S]*?<td[^>]*>\s*([^<]+?)\s*</td>`),
}

type Config struct {
	TTL time.Duration
	Now func() time.Time
}

type Prober struct {
	ttl time.Duration
	now func() time.Time

	mu       sync.Mutex
	cached   interfaces.BackendInfo
	cachedAt time.Time
}

type ipWhoIsResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	IP         string `json:"ip"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Connection struct {
		ISP string `json:"isp"`
	} `json:"connection"`
}

func New(cfg Config) *Prober {
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = defaultTTL
	}
	now := cfg.Now
	if now == nil {
		now = time.Now
	}
	return &Prober{
		ttl: ttl,
		now: now,
	}
}

func (p *Prober) Probe(ctx context.Context) (interfaces.BackendInfo, error) {
	if p == nil {
		return interfaces.BackendInfo{}, fmt.Errorf("backend prober is nil")
	}

	p.mu.Lock()
	if !p.cachedAt.IsZero() && p.now().Sub(p.cachedAt) < p.ttl {
		cached := p.cached
		p.mu.Unlock()
		return cached, backendProbeError(cached)
	}
	p.mu.Unlock()

	probeCtx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	info, err := p.probeOnce(probeCtx)

	p.mu.Lock()
	p.cached = info
	p.cachedAt = p.now()
	p.mu.Unlock()

	return info, err
}

func (p *Prober) probeOnce(ctx context.Context) (interfaces.BackendInfo, error) {
	ipInfo, err := lookupCurrentIP(ctx)
	if err != nil {
		return interfaces.BackendInfo{
			Source: "ipwho.is",
			Error:  err.Error(),
		}, err
	}

	location, locErr := lookupLocation(ctx, ipInfo.IP)
	backend := interfaces.BackendInfo{
		IP:        ipInfo.IP,
		UpdatedAt: p.now().UTC().Format(time.RFC3339),
	}

	if locErr == nil && location != "" {
		backend.Location = location
		backend.Source = "ip.cn"
		return backend, nil
	}

	backend.Location = buildFallbackLocation(ipInfo)
	backend.Source = "ipwho.is"
	if locErr != nil {
		backend.Error = locErr.Error()
	}
	return backend, locErr
}

func lookupCurrentIP(ctx context.Context) (ipWhoIsResponse, error) {
	resp, err := netx.RequestUnsafe(ctx, nil, interfaces.RequestOptions{
		Method: http.MethodGet,
		URL:    currentIPLookupURL,
		Headers: map[string]string{
			"Accept":     "application/json",
			"User-Agent": "sakiko/0.1",
		},
		Network: interfaces.ROptionsTCP,
	})
	if err != nil {
		return ipWhoIsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return ipWhoIsResponse{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var payload ipWhoIsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return ipWhoIsResponse{}, err
	}
	if !payload.Success || strings.TrimSpace(payload.IP) == "" {
		if payload.Message == "" {
			payload.Message = "ip lookup failed"
		}
		return ipWhoIsResponse{}, errors.New(payload.Message)
	}
	return payload, nil
}

func lookupLocation(ctx context.Context, ip string) (string, error) {
	if strings.TrimSpace(ip) == "" {
		return "", fmt.Errorf("empty backend ip")
	}

	resp, err := netx.RequestUnsafe(ctx, nil, interfaces.RequestOptions{
		Method: http.MethodGet,
		URL:    fmt.Sprintf(locationLookupFormat, ip),
		Headers: map[string]string{
			"Accept":     "text/html,application/xhtml+xml",
			"User-Agent": "sakiko/0.1",
		},
		Network: interfaces.ROptionsTCP,
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	decodedBody, err := decodeHTMLBody(resp.Header.Get("Content-Type"), body)
	if err != nil {
		return "", err
	}

	location, ok := extractLocationFromHTML(decodedBody)
	if !ok {
		return "", fmt.Errorf("backend location not found in ip.cn response")
	}
	if location == "" {
		return "", fmt.Errorf("backend location is empty")
	}
	return location, nil
}

func decodeHTMLBody(contentType string, body []byte) (string, error) {
	reader, err := charset.NewReader(bytes.NewReader(body), contentType)
	if err != nil {
		return string(body), nil
	}

	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func extractLocationFromHTML(raw string) (string, bool) {
	text := html.UnescapeString(raw)
	for _, pattern := range locationPatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) < 2 {
			continue
		}

		location := strings.Join(strings.Fields(matches[1]), " ")
		if location != "" {
			return location, true
		}
	}
	return "", false
}

func buildFallbackLocation(info ipWhoIsResponse) string {
	parts := make([]string, 0, 3)
	if country := strings.TrimSpace(info.Country); country != "" {
		parts = append(parts, country)
	}
	if city := strings.TrimSpace(info.City); city != "" {
		parts = append(parts, city)
	}
	if isp := strings.TrimSpace(info.Connection.ISP); isp != "" {
		parts = append(parts, isp)
	}
	return strings.Join(parts, " ")
}

func withDefaultTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, defaultProbeTimeout)
}

func backendProbeError(info interfaces.BackendInfo) error {
	if strings.TrimSpace(info.Error) == "" {
		return nil
	}
	return errors.New(info.Error)
}
