package geo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"
	"sakiko.local/sakiko-core/netx"

	mihomoresolver "github.com/metacubex/mihomo/component/resolver"
	"go.uber.org/zap"
)

var (
	outboundLookupURL  = "https://ipwho.is/"
	ipLookupURLPattern = "https://ipwho.is/%s"
	lookupInboundFunc  = lookupInbound
	lookupOutboundFunc = lookupOutbound
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

func (m *Macro) Type() interfaces.MacroType {
	return interfaces.MacroGeo
}

func (m *Macro) Run(proxy interfaces.Vendor, task *interfaces.Task) error {
	timeout := time.Duration(task.Config.Normalize().TaskTimeoutMillis) * time.Millisecond
	if timeout <= 0 {
		timeout = 6 * time.Second
	}
	attempts := geoRetryAttempts(task)

	inbound, inboundErr := retryGeoLookup("inbound", proxy, timeout, attempts, lookupInboundFunc)
	outbound, outboundErr := retryGeoLookup("outbound", proxy, timeout, attempts, lookupOutboundFunc)

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
	name string,
	proxy interfaces.Vendor,
	timeout time.Duration,
	attempts int,
	run func(interfaces.Vendor, time.Duration) (interfaces.GeoIPInfo, error),
) (interfaces.GeoIPInfo, error) {
	if attempts < 1 {
		attempts = 1
	}

	var lastInfo interfaces.GeoIPInfo
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		info, err := run(proxy, timeout)
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
		time.Sleep(150 * time.Millisecond)
	}

	return lastInfo, lastErr
}

func lookupInbound(proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	if proxy == nil {
		return interfaces.GeoIPInfo{}, fmt.Errorf("proxy is nil")
	}

	host := extractHost(proxy.ProxyInfo().Address)
	if host == "" {
		return interfaces.GeoIPInfo{}, fmt.Errorf("empty inbound address")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ip, err := resolveHost(ctx, host)
	if err != nil {
		return interfaces.GeoIPInfo{Address: host}, err
	}

	info, err := lookupIPInfo(proxy, ip, timeout)
	info.Address = host
	return info, err
}

func lookupOutbound(proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	return lookupCurrentIPInfo(proxy, timeout)
}

func lookupCurrentIPInfo(proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	return requestIPInfo(proxy, outboundLookupURL, timeout)
}

func lookupIPInfo(proxy interfaces.Vendor, ip string, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	url := fmt.Sprintf(ipLookupURLPattern, ip)
	return requestIPInfo(proxy, url, timeout)
}

func requestIPInfo(proxy interfaces.Vendor, url string, timeout time.Duration) (interfaces.GeoIPInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
