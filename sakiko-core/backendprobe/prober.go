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
	defaultTTL          = 10 * time.Minute
	defaultProbeTimeout = 4 * time.Second
	currentIPLookupURL  = "https://ipwho.is/"
	ipCNLookupURL       = "https://www.ip.cn/"
)

var (
	ipPattern         = regexp.MustCompile(`(?i)(?:^|[{"',\s])ip["']?\s*[:=]\s*["']?((?:\d{1,3}\.){3}\d{1,3})["']?`)
	ipCNFieldPatterns = map[string]*regexp.Regexp{
		"country":  regexp.MustCompile(`(?i)(?:^|[{"',\s])country["']?\s*[:=]\s*["']([^"']+)["']`),
		"province": regexp.MustCompile(`(?i)(?:^|[{"',\s])province["']?\s*[:=]\s*["']([^"']+)["']`),
		"city":     regexp.MustCompile(`(?i)(?:^|[{"',\s])city["']?\s*[:=]\s*["']([^"']+)["']`),
		"district": regexp.MustCompile(`(?i)(?:^|[{"',\s])district["']?\s*[:=]\s*["']([^"']+)["']`),
		"isp":      regexp.MustCompile(`(?i)(?:^|[{"',\s])isp["']?\s*[:=]\s*["']([^"']+)["']`),
	}
	locationPatterns = []*regexp.Regexp{
		regexp.MustCompile(`所在地理位置[\s\S]*?<td[^>]*>\s*([^<]+?)\s*</td>`),
		regexp.MustCompile(`归属地理位置[\s\S]*?<td[^>]*>\s*([^<]+?)\s*</td>`),
		regexp.MustCompile(`地理位置[\s\S]*?<td[^>]*>\s*([^<]+?)\s*</td>`),
	}
)

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

type ipCNCurrentInfo struct {
	IP       string
	Country  string
	Province string
	City     string
	District string
	ISP      string
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
	ipCNErr := error(nil)
	if backend, err := lookupCurrentBackendFromIPCN(ctx); err == nil {
		backend.UpdatedAt = p.now().UTC().Format(time.RFC3339)
		return backend, nil
	} else {
		ipCNErr = err
	}

	ipInfo, err := lookupCurrentIP(ctx)
	if err != nil {
		combinedError := joinProbeErrors(
			formatProbeError("ip.cn", ipCNErr),
			formatProbeError("ipwho.is", err),
		)
		return interfaces.BackendInfo{
			Source: "ip.cn -> ipwho.is",
			Error:  combinedError,
		}, errors.Join(ipCNErr, err)
	}

	return interfaces.BackendInfo{
		IP:        ipInfo.IP,
		Location:  buildFallbackLocation(ipInfo),
		Source:    "ipwho.is",
		UpdatedAt: p.now().UTC().Format(time.RFC3339),
	}, nil
}

func lookupCurrentBackendFromIPCN(ctx context.Context) (interfaces.BackendInfo, error) {
	resp, err := netx.RequestUnsafe(ctx, nil, interfaces.RequestOptions{
		Method: http.MethodGet,
		URL:    ipCNLookupURL,
		Headers: map[string]string{
			"Accept":     "text/html,application/xhtml+xml",
			"User-Agent": "sakiko/0.1",
		},
		Network: interfaces.ROptionsTCP,
	})
	if err != nil {
		return interfaces.BackendInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return interfaces.BackendInfo{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return interfaces.BackendInfo{}, err
	}

	decodedBody, err := decodeHTMLBody(resp.Header.Get("Content-Type"), body)
	if err != nil {
		return interfaces.BackendInfo{}, err
	}

	info, err := extractCurrentInfoFromIPCN(decodedBody)
	if err != nil {
		return interfaces.BackendInfo{}, err
	}

	location := buildIPCNLocation(info)
	if strings.TrimSpace(info.IP) == "" && location == "" {
		return interfaces.BackendInfo{}, fmt.Errorf("backend info not found in ip.cn response")
	}

	return interfaces.BackendInfo{
		IP:       strings.TrimSpace(info.IP),
		Location: location,
		Source:   "ip.cn",
	}, nil
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

func extractCurrentInfoFromIPCN(raw string) (ipCNCurrentInfo, error) {
	text := html.UnescapeString(raw)

	info := ipCNCurrentInfo{
		IP:       extractPatternValue(ipPattern, text),
		Country:  extractPatternValue(ipCNFieldPatterns["country"], text),
		Province: extractPatternValue(ipCNFieldPatterns["province"], text),
		City:     extractPatternValue(ipCNFieldPatterns["city"], text),
		District: extractPatternValue(ipCNFieldPatterns["district"], text),
		ISP:      extractPatternValue(ipCNFieldPatterns["isp"], text),
	}

	if buildIPCNLocation(info) != "" || info.IP != "" {
		return info, nil
	}

	if location, ok := extractLocationFromHTML(text); ok {
		info.LocationTokens(location)
		if buildIPCNLocation(info) != "" {
			return info, nil
		}
	}

	return ipCNCurrentInfo{}, fmt.Errorf("backend info not found in ip.cn response")
}

func (i *ipCNCurrentInfo) LocationTokens(location string) {
	parts := strings.Fields(strings.TrimSpace(location))
	if len(parts) == 0 {
		return
	}
	if i.Country == "" && len(parts) > 0 {
		i.Country = parts[0]
	}
	if i.Province == "" && len(parts) > 1 {
		i.Province = parts[1]
	}
	if i.City == "" && len(parts) > 2 {
		i.City = parts[2]
	}
	if i.District == "" && len(parts) > 3 {
		i.District = parts[3]
	}
	if i.ISP == "" && len(parts) > 4 {
		i.ISP = strings.Join(parts[4:], " ")
	}
}

func extractPatternValue(pattern *regexp.Regexp, text string) string {
	if pattern == nil {
		return ""
	}
	matches := pattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
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

func buildIPCNLocation(info ipCNCurrentInfo) string {
	parts := make([]string, 0, 5)
	appendUnique := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		for _, existing := range parts {
			if existing == value {
				return
			}
		}
		parts = append(parts, value)
	}

	appendUnique(info.Country)
	appendUnique(info.Province)
	appendUnique(info.City)
	appendUnique(info.District)
	appendUnique(info.ISP)
	return strings.Join(parts, " ")
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

func formatProbeError(source string, err error) string {
	source = strings.TrimSpace(source)
	if err == nil {
		return ""
	}
	if source == "" {
		return err.Error()
	}
	return source + ": " + err.Error()
}

func joinProbeErrors(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		nonEmpty = append(nonEmpty, part)
	}
	return strings.Join(nonEmpty, " | ")
}
