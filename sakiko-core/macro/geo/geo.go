package geo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"
	"sakiko.local/sakiko-core/netx"

	mihomoresolver "github.com/metacubex/mihomo/component/resolver"
	"go.uber.org/zap"
)

var (
	outboundLookupURL         = "https://ipwho.is/"
	ipLookupURLPattern        = "https://ipwho.is/%s"
	outboundLookupURLIPSB     = "https://api.ip.sb/geoip"
	ipLookupURLPatternIPSB    = "https://api.ip.sb/geoip/%s"
	outboundLookupURLIPAPICo  = "https://ipapi.co/json/"
	ipLookupURLPatternIPAPICo = "https://ipapi.co/%s/json/"
	lookupInboundFunc         = lookupInbound
	lookupOutboundFunc        = lookupOutbound
)

type Macro struct {
	Inbound  interfaces.GeoIPInfo
	Outbound interfaces.GeoIPInfo
}

type ipWhoIsResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	IP          string `json:"ip"`
	City        string `json:"city"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Connection  struct {
		ASN int    `json:"asn"`
		Org string `json:"org"`
		ISP string `json:"isp"`
	} `json:"connection"`
}

type ipSBResponse struct {
	IP           string `json:"ip"`
	CountryCode  string `json:"country_code"`
	Country      string `json:"country"`
	City         string `json:"city"`
	ASN          any    `json:"asn"`
	Organization string `json:"organization"`
}

type ipAPICoResponse struct {
	IP          string `json:"ip"`
	City        string `json:"city"`
	Country     string `json:"country_name"`
	CountryCode string `json:"country_code"`
	ASN         any    `json:"asn"`
	Org         string `json:"org"`
	Error       bool   `json:"error"`
	Reason      string `json:"reason"`
}

type geoLookupProvider struct {
	Name   string
	URL    func(targetIP string) string
	Decode func(*http.Response) (interfaces.GeoIPInfo, error)
}

func (m *Macro) Type() interfaces.MacroType {
	return interfaces.MacroGeo
}

func (m *Macro) Run(ctx context.Context, proxy interfaces.Vendor, task *interfaces.Task) error {
	if ctx == nil {
		ctx = context.Background()
	}
	timeout := time.Duration(task.Config.Normalize().TaskTimeoutMillis) * time.Millisecond
	if timeout <= 0 {
		timeout = 6 * time.Second
	}
	attempts := geoRetryAttempts(task)

	inbound, inboundErr := retryGeoLookup(ctx, "inbound", proxy, timeout, attempts, lookupInboundFunc)
	outbound, outboundErr := retryGeoLookup(ctx, "outbound", proxy, timeout, attempts, lookupOutboundFunc)

	m.Inbound = inbound
	m.Outbound = outbound

	if inboundErr != nil {
		m.Inbound.Error = inboundErr.Error()
	}
	if outboundErr != nil {
		m.Outbound.Error = outboundErr.Error()
	}
	if inboundErr != nil && outboundErr != nil {
		return errors.Join(
			fmt.Errorf("inbound geo lookup failed: %w", inboundErr),
			fmt.Errorf("outbound geo lookup failed: %w", outboundErr),
		)
	}
	return nil
}

func geoRetryAttempts(task *interfaces.Task) int {
	if task == nil {
		return 1
	}
	attempts := int(task.Config.Normalize().TaskRetry)
	if attempts < 1 {
		return 1
	}
	return attempts
}

func retryGeoLookup(
	ctx context.Context,
	name string,
	proxy interfaces.Vendor,
	timeout time.Duration,
	attempts int,
	run func(context.Context, interfaces.Vendor, time.Duration) (interfaces.GeoIPInfo, error),
) (interfaces.GeoIPInfo, error) {
	if attempts < 1 {
		attempts = 1
	}

	var lastInfo interfaces.GeoIPInfo
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return lastInfo, err
		}
		info, err := run(ctx, proxy, timeout)
		if err == nil {
			return info, nil
		}

		lastInfo = info
		lastErr = err

		if attempt >= attempts {
			break
		}

		geoLogger().Info("retrying geo lookup",
			zap.String("stage", name),
			zap.Int("attempt", attempt+1),
			zap.Int("max_attempts", attempts),
			zap.Duration("timeout", timeout),
			zap.String("error", err.Error()),
		)
		select {
		case <-ctx.Done():
			return lastInfo, ctx.Err()
		case <-time.After(150 * time.Millisecond):
		}
	}

	return lastInfo, lastErr
}

func lookupInbound(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	if proxy == nil {
		return interfaces.GeoIPInfo{}, fmt.Errorf("proxy is nil")
	}

	host := extractHost(proxy.ProxyInfo().Address)
	if host == "" {
		return interfaces.GeoIPInfo{}, fmt.Errorf("empty inbound address")
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ip, err := resolveHost(ctx, host)
	if err != nil {
		return interfaces.GeoIPInfo{Address: host}, err
	}

	info, err := lookupIPInfo(ctx, proxy, ip, timeout)
	info.Address = host
	return info, err
}

func lookupOutbound(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	return lookupCurrentIPInfo(ctx, proxy, timeout)
}

func lookupCurrentIPInfo(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	return requestIPInfo(ctx, proxy, "", timeout)
}

func lookupIPInfo(ctx context.Context, proxy interfaces.Vendor, ip string, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	return requestIPInfo(ctx, proxy, ip, timeout)
}

func requestIPInfo(ctx context.Context, proxy interfaces.Vendor, targetIP string, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	providers := geoLookupProviders()
	if len(providers) == 0 {
		return interfaces.GeoIPInfo{}, fmt.Errorf("no geo providers configured")
	}

	deadline := time.Now().Add(timeout)
	errs := make([]error, 0, len(providers))
	for i, provider := range providers {
		if err := ctx.Err(); err != nil {
			return interfaces.GeoIPInfo{}, err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			errs = append(errs, fmt.Errorf("%s: timeout budget exhausted", provider.Name))
			break
		}

		slotsLeft := len(providers) - i
		providerTimeout := remaining / time.Duration(slotsLeft)
		if providerTimeout <= 0 || providerTimeout > remaining {
			providerTimeout = remaining
		}

		info, err := requestIPInfoFromProvider(ctx, proxy, provider.URL(targetIP), providerTimeout, provider.Decode)
		if err == nil {
			return info, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", provider.Name, err))
	}

	if len(errs) == 1 {
		return interfaces.GeoIPInfo{}, errs[0]
	}
	return interfaces.GeoIPInfo{}, fmt.Errorf("all geo providers failed: %w", errors.Join(errs...))
}

func requestIPInfoFromProvider(
	ctx context.Context,
	proxy interfaces.Vendor,
	url string,
	timeout time.Duration,
	decode func(*http.Response) (interfaces.GeoIPInfo, error),
) (interfaces.GeoIPInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := netx.RequestUnsafe(ctx, proxy, interfaces.RequestOptions{
		Method: http.MethodGet,
		URL:    url,
		Headers: map[string]string{
			"Accept":     "application/json",
			"User-Agent": "sakiko/0.1",
		},
		Network: interfaces.ROptionsTCP,
	})
	if err != nil {
		return interfaces.GeoIPInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return interfaces.GeoIPInfo{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return decode(resp)
}

func decodeIPWhoIsResponse(resp *http.Response) (interfaces.GeoIPInfo, error) {
	var payload ipWhoIsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return interfaces.GeoIPInfo{}, err
	}
	if !payload.Success && payload.IP == "" {
		if payload.Message == "" {
			payload.Message = "ip lookup failed"
		}
		return interfaces.GeoIPInfo{}, errors.New(payload.Message)
	}

	return interfaces.GeoIPInfo{
		Address:        payload.IP,
		IP:             payload.IP,
		ASN:            payload.Connection.ASN,
		ASOrganization: payload.Connection.Org,
		ISP:            payload.Connection.ISP,
		Country:        payload.Country,
		City:           payload.City,
		CountryCode:    payload.CountryCode,
	}, nil
}

func decodeIPSBResponse(resp *http.Response) (interfaces.GeoIPInfo, error) {
	var payload ipSBResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return interfaces.GeoIPInfo{}, err
	}
	if strings.TrimSpace(payload.IP) == "" {
		return interfaces.GeoIPInfo{}, fmt.Errorf("ip.sb lookup failed")
	}

	org := strings.TrimSpace(payload.Organization)
	return interfaces.GeoIPInfo{
		Address:        payload.IP,
		IP:             payload.IP,
		ASN:            parseASN(payload.ASN),
		ASOrganization: org,
		ISP:            org,
		Country:        payload.Country,
		City:           payload.City,
		CountryCode:    payload.CountryCode,
	}, nil
}

func decodeIPAPICoResponse(resp *http.Response) (interfaces.GeoIPInfo, error) {
	var payload ipAPICoResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return interfaces.GeoIPInfo{}, err
	}
	if payload.Error && strings.TrimSpace(payload.IP) == "" {
		reason := strings.TrimSpace(payload.Reason)
		if reason == "" {
			reason = "ipapi.co lookup failed"
		}
		return interfaces.GeoIPInfo{}, errors.New(reason)
	}
	if strings.TrimSpace(payload.IP) == "" {
		return interfaces.GeoIPInfo{}, fmt.Errorf("ipapi.co lookup failed")
	}

	org := strings.TrimSpace(payload.Org)
	return interfaces.GeoIPInfo{
		Address:        payload.IP,
		IP:             payload.IP,
		ASN:            parseASN(payload.ASN),
		ASOrganization: org,
		ISP:            org,
		Country:        payload.Country,
		City:           payload.City,
		CountryCode:    payload.CountryCode,
	}, nil
}

func geoLookupProviders() []geoLookupProvider {
	return []geoLookupProvider{
		{
			Name: "ipwho.is",
			URL: func(targetIP string) string {
				if strings.TrimSpace(targetIP) == "" {
					return outboundLookupURL
				}
				return fmt.Sprintf(ipLookupURLPattern, targetIP)
			},
			Decode: decodeIPWhoIsResponse,
		},
		{
			Name: "ip.sb",
			URL: func(targetIP string) string {
				if strings.TrimSpace(targetIP) == "" {
					return outboundLookupURLIPSB
				}
				return fmt.Sprintf(ipLookupURLPatternIPSB, targetIP)
			},
			Decode: decodeIPSBResponse,
		},
		{
			Name: "ipapi.co",
			URL: func(targetIP string) string {
				if strings.TrimSpace(targetIP) == "" {
					return outboundLookupURLIPAPICo
				}
				return fmt.Sprintf(ipLookupURLPatternIPAPICo, targetIP)
			},
			Decode: decodeIPAPICoResponse,
		},
	}
}

func parseASN(value any) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case float32:
		return int(typed)
	case int:
		return typed
	case int64:
		return int(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return int(parsed)
		}
	case string:
		raw := strings.TrimSpace(strings.ToUpper(typed))
		raw = strings.TrimPrefix(raw, "AS")
		parsed, err := strconv.Atoi(raw)
		if err == nil {
			return parsed
		}
	}
	return 0
}

func resolveHost(ctx context.Context, host string) (string, error) {
	if ip := net.ParseIP(host); ip != nil {
		return ip.String(), nil
	}

	if ips, err := mihomoresolver.LookupIPWithResolver(ctx, host, mihomoresolver.ProxyServerHostResolver); err == nil {
		for _, ip := range ips {
			if ip.Is4() {
				return ip.String(), nil
			}
		}
		if len(ips) > 0 {
			return ips[0].String(), nil
		}
	}

	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return "", err
	}
	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String(), nil
		}
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no ip address resolved for %s", host)
	}
	return ips[0].String(), nil
}

func extractHost(address string) string {
	address = strings.TrimSpace(address)
	if address == "" || strings.EqualFold(address, "direct") {
		return ""
	}

	if host, port, err := net.SplitHostPort(address); err == nil && port != "" {
		return strings.Trim(host, "[]")
	}

	trimmed := strings.Trim(address, "[]")
	if ip := net.ParseIP(trimmed); ip != nil {
		return trimmed
	}

	if strings.Count(address, ":") == 1 {
		if idx := strings.LastIndex(address, ":"); idx > 0 {
			return address[:idx]
		}
	}

	return trimmed
}

func geoLogger() *zap.Logger {
	return logx.Named("core.macro.geo")
}
