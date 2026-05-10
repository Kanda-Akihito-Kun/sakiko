package netenv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/netx"
)

const (
	defaultProbeTimeout      = 6 * time.Second
	outboundLookupURL        = "https://ipwho.is/"
	outboundLookupURLIPSB    = "https://api.ip.sb/geoip"
	outboundLookupURLIPAPICo = "https://ipapi.co/json/"
)

type probeProvider struct {
	Name   string
	URL    string
	Decode func(*http.Response) (interfaces.BackendInfo, error)
}

type ipWhoIsResponse struct {
	Success     bool   `json:"success"`
	IP          string `json:"ip"`
	City        string `json:"city"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
	Message     string `json:"message"`
	Connection  struct {
		Org string `json:"org"`
		ISP string `json:"isp"`
	} `json:"connection"`
}

type ipSBResponse struct {
	IP           string `json:"ip"`
	City         string `json:"city"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
	Organization string `json:"organization"`
}

type ipAPICoResponse struct {
	IP          string `json:"ip"`
	City        string `json:"city"`
	Country     string `json:"country_name"`
	CountryCode string `json:"country_code"`
	Org         string `json:"org"`
	Error       bool   `json:"error"`
	Reason      string `json:"reason"`
}

func Probe(ctx context.Context) interfaces.BackendInfo {
	info, err := probeWithTimeout(ctx, defaultProbeTimeout)
	if err != nil {
		return interfaces.BackendInfo{
			Source:    "network-env",
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			Error:     err.Error(),
		}
	}
	return info
}

func probeWithTimeout(ctx context.Context, timeout time.Duration) (interfaces.BackendInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	providers := probeProviders()
	deadline := time.Now().Add(timeout)
	errs := make([]error, 0, len(providers))
	for i, provider := range providers {
		if err := ctx.Err(); err != nil {
			return interfaces.BackendInfo{}, err
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

		info, err := requestProvider(ctx, provider, providerTimeout)
		if err == nil {
			info.Source = provider.Name
			info.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			return info, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", provider.Name, err))
	}

	if len(errs) == 0 {
		return interfaces.BackendInfo{}, fmt.Errorf("no network env provider configured")
	}
	if len(errs) == 1 {
		return interfaces.BackendInfo{}, errs[0]
	}
	return interfaces.BackendInfo{}, fmt.Errorf("all network env providers failed: %w", errors.Join(errs...))
}

func requestProvider(ctx context.Context, provider probeProvider, timeout time.Duration) (interfaces.BackendInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := netx.RequestUnsafe(ctx, nil, interfaces.RequestOptions{
		Method: http.MethodGet,
		URL:    provider.URL,
		Headers: map[string]string{
			"Accept":     "application/json",
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

	return provider.Decode(resp)
}

func decodeIPWhoIs(resp *http.Response) (interfaces.BackendInfo, error) {
	var payload ipWhoIsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return interfaces.BackendInfo{}, err
	}
	if !payload.Success && strings.TrimSpace(payload.IP) == "" {
		reason := strings.TrimSpace(payload.Message)
		if reason == "" {
			reason = "ipwho.is lookup failed"
		}
		return interfaces.BackendInfo{}, errors.New(reason)
	}

	return interfaces.BackendInfo{
		IP:       payload.IP,
		Location: formatLocation(payload.CountryCode, payload.Country, payload.City, payload.Connection.Org, payload.Connection.ISP),
	}, nil
}

func decodeIPSB(resp *http.Response) (interfaces.BackendInfo, error) {
	var payload ipSBResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return interfaces.BackendInfo{}, err
	}
	if strings.TrimSpace(payload.IP) == "" {
		return interfaces.BackendInfo{}, fmt.Errorf("ip.sb lookup failed")
	}

	return interfaces.BackendInfo{
		IP:       payload.IP,
		Location: formatLocation(payload.CountryCode, payload.Country, payload.City, payload.Organization, ""),
	}, nil
}

func decodeIPAPICo(resp *http.Response) (interfaces.BackendInfo, error) {
	var payload ipAPICoResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return interfaces.BackendInfo{}, err
	}
	if payload.Error && strings.TrimSpace(payload.IP) == "" {
		reason := strings.TrimSpace(payload.Reason)
		if reason == "" {
			reason = "ipapi.co lookup failed"
		}
		return interfaces.BackendInfo{}, errors.New(reason)
	}
	if strings.TrimSpace(payload.IP) == "" {
		return interfaces.BackendInfo{}, fmt.Errorf("ipapi.co lookup failed")
	}

	return interfaces.BackendInfo{
		IP:       payload.IP,
		Location: formatLocation(payload.CountryCode, payload.Country, payload.City, payload.Org, ""),
	}, nil
}

func probeProviders() []probeProvider {
	return []probeProvider{
		{Name: "ipwho.is", URL: outboundLookupURL, Decode: decodeIPWhoIs},
		{Name: "ip.sb", URL: outboundLookupURLIPSB, Decode: decodeIPSB},
		{Name: "ipapi.co", URL: outboundLookupURLIPAPICo, Decode: decodeIPAPICo},
	}
}

func formatLocation(countryCode string, country string, city string, org string, isp string) string {
	parts := make([]string, 0, 5)
	if value := strings.TrimSpace(countryCode); value != "" {
		parts = append(parts, strings.ToUpper(value))
	}
	if value := strings.TrimSpace(country); value != "" {
		parts = append(parts, value)
	}
	if value := strings.TrimSpace(city); value != "" {
		parts = append(parts, value)
	}
	if value := strings.TrimSpace(org); value != "" {
		parts = append(parts, value)
	}
	if value := strings.TrimSpace(isp); value != "" && !strings.EqualFold(strings.TrimSpace(org), value) {
		parts = append(parts, value)
	}
	return strings.Join(parts, " | ")
}
