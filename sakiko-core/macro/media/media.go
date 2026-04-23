package media

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"
	"sakiko.local/sakiko-core/netx"

	"go.uber.org/zap"
)

const (
	netflixHost              = "www.netflix.com"
	netflixTitleURL1         = "https://www.netflix.com/title/81280792"
	netflixTitleURL2         = "https://www.netflix.com/title/70143836"
	huluAuthHost             = "auth.hulu.com"
	huluAuthURL              = "https://auth.hulu.com/v4/web/password/authenticate"
	huluAuthCookie           = "_h_csrf_id=b0b3da20eccdc796dd61d9145a095be4927a2ff56821ad4d3f91804fd6f918ea"
	huluAuthBody             = "csrf=fdc1427eccde53326e27d7575c436595e28299dc420232ff26075ca06bbb28ed&password=Jam0.5cm~&scenario=web_password_login&user_email=me%40jamchoi.cc"
	bilibiliHKMCTWHost       = "api.bilibili.com"
	bilibiliHKMCTWURLPattern = "https://api.bilibili.com/pgc/player/web/playurl?avid=18281381&cid=29892777&qn=0&type=&otype=json&ep_id=183799&fourk=1&fnver=0&fnval=16&session=%s&module=bangumi"
	mediaProbeAttemptTimeout = 2 * time.Second
	modeProbeTimeout         = 8 * time.Second
	publicLookupTimeout      = 4 * time.Second
	maxModeProbeIPs          = 2
	mediaProbeRetryAttempts  = 2
	mediaProbeConcurrency    = 5
	maxMediaProbeTimeout     = 6 * time.Second
)

var (
	mediaBrowserUA       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36 Edg/112.0.1722.64"
	netflixContextRegexp = regexp.MustCompile(`(?s)netflix\.reactContext\s*=\s*(\{.*?\});`)
	doHQueryPatterns     = []string{
		"https://cloudflare-dns.com/dns-query?name=%s&type=%s",
		"https://dns.google/resolve?name=%s&type=%s",
	}
)

type Macro struct {
	Result interfaces.MediaUnlockResult
}

type httpSnapshot struct {
	StatusCode int
	FinalURL   string
	Headers    http.Header
	Body       []byte
}

type doHResponse struct {
	Answer []struct {
		Data string `json:"data"`
	} `json:"Answer"`
}

type requestSpec struct {
	Method        string
	URL           string
	Headers       map[string]string
	Body          []byte
	Host          string
	TLSServerName string
	NoRedir       bool
}

type mediaProbeSpec struct {
	name string
	run  func(context.Context, interfaces.Vendor) interfaces.MediaUnlockPlatformResult
}

func (m *Macro) Type() interfaces.MacroType {
	return interfaces.MacroMedia
}

func (m *Macro) Run(ctx context.Context, proxy interfaces.Vendor, task *interfaces.Task) error {
	if ctx == nil {
		ctx = context.Background()
	}
	probeTimeout := resolveProbeTimeout(task)
	probes := []mediaProbeSpec{
		{name: "ChatGPT", run: probeChatGPT},
		{name: "Claude", run: probeClaude},
		{name: "Gemini", run: probeGemini},
		{name: "YouTube Premium", run: probeYouTubePremium},
		{name: "Netflix", run: probeNetflix},
		{name: "Hulu", run: probeHulu},
		{name: "Prime Video", run: probePrimeVideo},
		{name: "HBO Max", run: probeHBOMax},
		{name: "Bilibili HK/MO/TW", run: probeBilibiliHKMCTW},
		{name: "Bilibili Taiwan", run: probeBilibiliTW},
		{name: "Abema", run: probeAbema},
		{name: "TikTok", run: probeTikTok},
	}
	results := make([]interfaces.MediaUnlockPlatformResult, len(probes))
	taskID := ""
	if task != nil {
		taskID = strings.TrimSpace(task.ID)
	}
	proxyInfo := interfaces.ProxyInfo{}
	if proxy != nil {
		proxyInfo = proxy.ProxyInfo()
	}
	logger := mediaLogger().With(
		zap.String("task_id", taskID),
		zap.String("proxy_name", proxyInfo.Name),
		zap.String("proxy_addr", proxyInfo.Address),
	)

	sem := make(chan struct{}, mediaProbeConcurrency)
	var wg sync.WaitGroup
	for index, probe := range probes {
		wg.Add(1)
		go func(i int, spec mediaProbeSpec) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				results[i] = finalizeResult(interfaces.MediaUnlockPlatformResult{
					Platform: interfaces.MediaUnlockPlatform(strings.ToLower(strings.ReplaceAll(spec.name, " ", "_"))),
					Name:     spec.name,
					Status:   interfaces.MediaUnlockStatusFailed,
					Error:    ctx.Err().Error(),
				})
				return
			case sem <- struct{}{}:
			}
			defer func() {
				<-sem
			}()
			startedAt := time.Now()
			results[i] = runProbeWithRetry(ctx, proxy, probeTimeout, spec.run)
			logger.Info("media probe finished",
				zap.String("platform", spec.name),
				zap.Duration("elapsed", time.Since(startedAt)),
				zap.String("status", string(results[i].Status)),
				zap.String("region", results[i].Region),
				zap.String("mode", string(results[i].Mode)),
				zap.String("error", results[i].Error),
			)
		}(index, probe)
	}
	wg.Wait()

	m.Result = interfaces.MediaUnlockResult{
		Items: results,
	}

	// Media probing is auxiliary data and must not poison node execution semantics.
	return nil
}

func runProbeWithRetry(
	ctx context.Context,
	proxy interfaces.Vendor,
	timeout time.Duration,
	run func(context.Context, interfaces.Vendor) interfaces.MediaUnlockPlatformResult,
) interfaces.MediaUnlockPlatformResult {
	first := runProbeAttempt(ctx, proxy, timeout, run)
	if first.Status != interfaces.MediaUnlockStatusFailed {
		return first
	}

	result := first
	for attempt := 2; attempt <= mediaProbeRetryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			result.Error = joinErrorTexts(result.Error, err.Error())
			return finalizeResult(result)
		}
		candidate := runProbeAttempt(ctx, proxy, timeout, run)
		if candidate.Status != interfaces.MediaUnlockStatusFailed {
			return candidate
		}
		candidate.Error = joinErrorTexts(result.Error, candidate.Error)
		result = finalizeResult(candidate)
	}

	return result
}

func runProbeAttempt(
	ctx context.Context,
	proxy interfaces.Vendor,
	timeout time.Duration,
	run func(context.Context, interfaces.Vendor) interfaces.MediaUnlockPlatformResult,
) interfaces.MediaUnlockPlatformResult {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return run(ctx, proxy)
}

func probeNetflix(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := interfaces.MediaUnlockPlatformResult{
		Platform: interfaces.MediaUnlockPlatformNetflix,
		Name:     "Netflix",
		Mode:     interfaces.MediaUnlockModeUnknown,
	}

	type netflixProbeResult struct {
		snapshot httpSnapshot
		err      error
	}

	specs := []requestSpec{
		{
			URL: netflixTitleURL1,
			Headers: map[string]string{
				"User-Agent": mediaBrowserUA,
			},
		},
		{
			URL: netflixTitleURL2,
			Headers: map[string]string{
				"User-Agent": mediaBrowserUA,
			},
		},
	}

	results := make([]netflixProbeResult, len(specs))
	var wg sync.WaitGroup
	for index, spec := range specs {
		wg.Add(1)
		go func(i int, req requestSpec) {
			defer wg.Done()
			results[i].snapshot, results[i].err = performRequest(ctx, proxy, req)
		}(index, spec)
	}
	wg.Wait()

	accessible := false
	for _, entry := range results {
		if entry.err == nil && accessibleNetflix(entry.snapshot) {
			accessible = true
		}
		if result.Region == "" {
			result.Region = entry.snapshot.netflixRegion()
		}
	}

	if results[0].err != nil && results[1].err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = joinErrors(results[0].err, results[1].err)
		return finalizeResult(result)
	}

	if accessible {
		result.Status = interfaces.MediaUnlockStatusYes
		result.Mode = inferNetflixUnlockMode(ctx, proxy)
		return finalizeResult(result)
	}

	result.Status = interfaces.MediaUnlockStatusOriginalsOnly
	return finalizeResult(result)
}

func probeHulu(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := interfaces.MediaUnlockPlatformResult{
		Platform: interfaces.MediaUnlockPlatformHulu,
		Name:     "Hulu",
		Mode:     interfaces.MediaUnlockModeUnknown,
	}

	snapshot, err := performRequest(ctx, proxy, buildHuluRequest(huluAuthURL, "", false))
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	status, errName, parseErr := evaluateHuluSnapshot(snapshot)
	if parseErr != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = parseErr.Error()
		return finalizeResult(result)
	}

	result.Status = status
	result.Error = errName
	if result.Status == interfaces.MediaUnlockStatusYes {
		result.Region = "US"
		result.Error = ""
		result.Mode = inferHuluUnlockMode(ctx, proxy)
	}
	if result.Status == interfaces.MediaUnlockStatusNo {
		result.Error = ""
	}
	return finalizeResult(result)
}

func probeBilibiliHKMCTW(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
	result := interfaces.MediaUnlockPlatformResult{
		Platform: interfaces.MediaUnlockPlatformBilibiliHMT,
		Name:     "Bilibili HK/MO/TW",
		Mode:     interfaces.MediaUnlockModeUnknown,
	}

	sessionID, err := randomHex(16)
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	requestURL := fmt.Sprintf(bilibiliHKMCTWURLPattern, sessionID)
	snapshot, err := performRequest(ctx, proxy, buildBilibiliRequest(requestURL, "", false))
	if err != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = err.Error()
		return finalizeResult(result)
	}

	status, parseErr := evaluateBilibiliSnapshot(snapshot)
	if parseErr != nil {
		result.Status = interfaces.MediaUnlockStatusFailed
		result.Error = parseErr.Error()
		return finalizeResult(result)
	}

	result.Status = status
	if result.Status == interfaces.MediaUnlockStatusYes {
		result.Region = "HK/MO/TW"
		result.Mode = inferBilibiliUnlockMode(ctx, proxy, requestURL)
	}
	return finalizeResult(result)
}

func performRequest(ctx context.Context, proxy interfaces.Vendor, spec requestSpec) (httpSnapshot, error) {
	snapshot := httpSnapshot{}
	resp, err := netx.RequestUnsafe(ctx, proxy, interfaces.RequestOptions{
		Method:        spec.Method,
		URL:           spec.URL,
		Headers:       spec.Headers,
		Body:          spec.Body,
		Host:          spec.Host,
		TLSServerName: spec.TLSServerName,
		NoRedir:       spec.NoRedir,
		Network:       interfaces.ROptionsTCP,
	})
	if err != nil {
		return httpSnapshot{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return httpSnapshot{}, err
	}

	snapshot.StatusCode = resp.StatusCode
	snapshot.Headers = resp.Header.Clone()
	if resp.Request != nil && resp.Request.URL != nil {
		snapshot.FinalURL = resp.Request.URL.String()
	}
	snapshot.Body = body
	return snapshot, nil
}

func accessibleNetflix(snapshot httpSnapshot) bool {
	return strings.Contains(string(snapshot.Body), "og:video")
}

func (s httpSnapshot) netflixRegion() string {
	matches := netflixContextRegexp.FindSubmatch(s.Body)
	if len(matches) < 2 {
		return ""
	}

	var payload struct {
		Models struct {
			Geo struct {
				Data struct {
					RequestCountry struct {
						ID string `json:"id"`
					} `json:"requestCountry"`
				} `json:"data"`
			} `json:"geo"`
		} `json:"models"`
	}
	if err := json.Unmarshal(matches[1], &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Models.Geo.Data.RequestCountry.ID)
}

func inferNetflixUnlockMode(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockMode {
	publicIPs, err := lookupPublicHostIPs(ctx, netflixHost)
	if err != nil || len(publicIPs) == 0 {
		return interfaces.MediaUnlockModeUnknown
	}

	attempted := 0
	directOnlyMisses := 0
	for _, ip := range limitStrings(publicIPs, maxModeProbeIPs) {
		attemptCtx, cancel := context.WithTimeout(ctx, modeProbeTimeout)
		first, err1 := performRequest(attemptCtx, proxy, buildNetflixRequest(netflixTitleURL1, ip, true))
		second, err2 := performRequest(attemptCtx, proxy, buildNetflixRequest(netflixTitleURL2, ip, true))
		cancel()

		if err1 != nil && err2 != nil {
			continue
		}

		attempted++
		if (err1 == nil && accessibleNetflix(first)) || (err2 == nil && accessibleNetflix(second)) {
			return interfaces.MediaUnlockModeNative
		}
		directOnlyMisses++
	}

	if attempted > 0 && directOnlyMisses == attempted {
		return interfaces.MediaUnlockModeDNS
	}
	return interfaces.MediaUnlockModeUnknown
}

func inferHuluUnlockMode(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockMode {
	publicIPs, err := lookupPublicHostIPs(ctx, huluAuthHost)
	if err != nil || len(publicIPs) == 0 {
		return interfaces.MediaUnlockModeUnknown
	}

	attempted := 0
	blockedCount := 0
	for _, ip := range limitStrings(publicIPs, maxModeProbeIPs) {
		attemptCtx, cancel := context.WithTimeout(ctx, modeProbeTimeout)
		snapshot, reqErr := performRequest(attemptCtx, proxy, buildHuluRequest(huluAuthURL, ip, true))
		cancel()
		if reqErr != nil {
			continue
		}

		status, _, parseErr := evaluateHuluSnapshot(snapshot)
		if parseErr != nil {
			continue
		}

		attempted++
		switch status {
		case interfaces.MediaUnlockStatusYes:
			return interfaces.MediaUnlockModeNative
		case interfaces.MediaUnlockStatusNo:
			blockedCount++
		}
	}

	if attempted > 0 && blockedCount == attempted {
		return interfaces.MediaUnlockModeDNS
	}
	return interfaces.MediaUnlockModeUnknown
}

func inferBilibiliUnlockMode(ctx context.Context, proxy interfaces.Vendor, requestURL string) interfaces.MediaUnlockMode {
	publicIPs, err := lookupPublicHostIPs(ctx, bilibiliHKMCTWHost)
	if err != nil || len(publicIPs) == 0 {
		return interfaces.MediaUnlockModeUnknown
	}

	attempted := 0
	blockedCount := 0
	for _, ip := range limitStrings(publicIPs, maxModeProbeIPs) {
		attemptCtx, cancel := context.WithTimeout(ctx, modeProbeTimeout)
		snapshot, reqErr := performRequest(attemptCtx, proxy, buildBilibiliRequest(requestURL, ip, true))
		cancel()
		if reqErr != nil {
			continue
		}

		status, parseErr := evaluateBilibiliSnapshot(snapshot)
		if parseErr != nil {
			continue
		}

		attempted++
		switch status {
		case interfaces.MediaUnlockStatusYes:
			return interfaces.MediaUnlockModeNative
		case interfaces.MediaUnlockStatusNo:
			blockedCount++
		}
	}

	if attempted > 0 && blockedCount == attempted {
		return interfaces.MediaUnlockModeDNS
	}
	return interfaces.MediaUnlockModeUnknown
}

func evaluateHuluSnapshot(snapshot httpSnapshot) (interfaces.MediaUnlockStatus, string, error) {
	var payload struct {
		Error struct {
			Name string `json:"name"`
		} `json:"error"`
	}
	if err := json.Unmarshal(snapshot.Body, &payload); err != nil {
		return interfaces.MediaUnlockStatusFailed, "", err
	}

	switch payload.Error.Name {
	case "LOGIN_FORBIDDEN":
		return interfaces.MediaUnlockStatusYes, "", nil
	case "GEO_BLOCKED":
		return interfaces.MediaUnlockStatusNo, "", nil
	case "":
		return interfaces.MediaUnlockStatusFailed, "page error", nil
	default:
		return interfaces.MediaUnlockStatusFailed, payload.Error.Name, nil
	}
}

func evaluateBilibiliSnapshot(snapshot httpSnapshot) (interfaces.MediaUnlockStatus, error) {
	var payload struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(snapshot.Body, &payload); err != nil {
		return interfaces.MediaUnlockStatusFailed, err
	}

	switch payload.Code {
	case 0:
		return interfaces.MediaUnlockStatusYes, nil
	case -10403:
		return interfaces.MediaUnlockStatusNo, nil
	default:
		return interfaces.MediaUnlockStatusFailed, fmt.Errorf("unexpected code: %d", payload.Code)
	}
}

func buildNetflixRequest(rawURL string, resolvedIP string, direct bool) requestSpec {
	spec := requestSpec{
		URL: rawURL,
		Headers: map[string]string{
			"User-Agent": mediaBrowserUA,
		},
	}
	if !direct {
		return spec
	}
	return applyResolvedIP(spec, netflixHost, resolvedIP)
}

func buildHuluRequest(rawURL string, resolvedIP string, direct bool) requestSpec {
	spec := requestSpec{
		Method: http.MethodPost,
		URL:    rawURL,
		Headers: map[string]string{
			"Cookie":       huluAuthCookie,
			"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
			"User-Agent":   mediaBrowserUA,
		},
		Body: []byte(huluAuthBody),
	}
	if !direct {
		return spec
	}
	return applyResolvedIP(spec, huluAuthHost, resolvedIP)
}

func buildBilibiliRequest(rawURL string, resolvedIP string, direct bool) requestSpec {
	spec := requestSpec{
		URL: rawURL,
		Headers: map[string]string{
			"Accept":     "application/json",
			"User-Agent": mediaBrowserUA,
		},
	}
	if !direct {
		return spec
	}
	return applyResolvedIP(spec, bilibiliHKMCTWHost, resolvedIP)
}

func applyResolvedIP(spec requestSpec, host string, resolvedIP string) requestSpec {
	if strings.TrimSpace(resolvedIP) == "" {
		return spec
	}

	replacedURL, err := replaceURLHost(spec.URL, resolvedIP)
	if err != nil {
		return spec
	}
	spec.URL = replacedURL
	spec.Host = host
	spec.TLSServerName = host
	return spec
}

func replaceURLHost(rawURL string, host string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if parsed.Port() != "" {
		parsed.Host = net.JoinHostPort(host, parsed.Port())
	} else if strings.Contains(host, ":") {
		parsed.Host = "[" + host + "]"
	} else {
		parsed.Host = host
	}
	return parsed.String(), nil
}

func lookupPublicHostIPs(ctx context.Context, host string) ([]string, error) {
	if strings.TrimSpace(host) == "" {
		return nil, fmt.Errorf("host is required")
	}

	ctx, cancel := context.WithTimeout(ctx, publicLookupTimeout)
	defer cancel()

	for _, pattern := range doHQueryPatterns {
		records, err := lookupDoHRecords(ctx, pattern, host)
		if err == nil && len(records) > 0 {
			return records, nil
		}
	}

	resolverCtx, resolverCancel := context.WithTimeout(ctx, publicLookupTimeout)
	defer resolverCancel()

	ips, err := net.DefaultResolver.LookupIPAddr(resolverCtx, host)
	if err != nil {
		return nil, err
	}

	records := make([]string, 0, len(ips))
	seen := map[string]struct{}{}
	for _, ip := range ips {
		value := strings.TrimSpace(ip.IP.String())
		if net.ParseIP(value) == nil {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		records = append(records, value)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no public records for %s", host)
	}
	return records, nil
}

func lookupDoHRecords(ctx context.Context, pattern string, host string) ([]string, error) {
	types := []string{"A", "AAAA"}
	records := make([]string, 0, 4)
	seen := map[string]struct{}{}

	for _, recordType := range types {
		queryURL := fmt.Sprintf(pattern, url.QueryEscape(host), recordType)
		resp, err := netx.RequestUnsafe(ctx, nil, interfaces.RequestOptions{
			Method: http.MethodGet,
			URL:    queryURL,
			Headers: map[string]string{
				"Accept":     "application/dns-json",
				"User-Agent": mediaBrowserUA,
			},
			Network: interfaces.ROptionsTCP,
		})
		if err != nil {
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			continue
		}

		var payload doHResponse
		if err := json.Unmarshal(body, &payload); err != nil {
			continue
		}
		for _, answer := range payload.Answer {
			ip := strings.TrimSpace(answer.Data)
			if net.ParseIP(ip) == nil {
				continue
			}
			if _, ok := seen[ip]; ok {
				continue
			}
			seen[ip] = struct{}{}
			records = append(records, ip)
		}
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no doh records for %s", host)
	}
	return records, nil
}

func randomHex(byteCount int) (string, error) {
	buf := make([]byte, byteCount)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func resolveProbeTimeout(task *interfaces.Task) time.Duration {
	timeout := mediaProbeAttemptTimeout
	if task != nil {
		cfg := task.Config.Normalize()
		if cfg.TaskTimeoutMillis > 0 {
			timeout = time.Duration(cfg.TaskTimeoutMillis) * time.Millisecond
		}
	}
	if timeout < mediaProbeAttemptTimeout {
		timeout = mediaProbeAttemptTimeout
	}
	if timeout > maxMediaProbeTimeout {
		timeout = maxMediaProbeTimeout
	}
	return timeout
}

func limitStrings(values []string, limit int) []string {
	if limit <= 0 || len(values) <= limit {
		return values
	}
	return values[:limit]
}

func joinErrors(errs ...error) string {
	parts := make([]string, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		message := strings.TrimSpace(err.Error())
		if message == "" {
			continue
		}
		duplicate := false
		for _, existing := range parts {
			if existing == message {
				duplicate = true
				break
			}
		}
		if !duplicate {
			parts = append(parts, message)
		}
	}
	return strings.Join(parts, " | ")
}

func joinErrorTexts(values ...string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		duplicate := false
		for _, existing := range parts {
			if existing == trimmed {
				duplicate = true
				break
			}
		}
		if !duplicate {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, " | ")
}

func mediaLogger() *zap.Logger {
	return logx.Named("core.macro.media")
}
