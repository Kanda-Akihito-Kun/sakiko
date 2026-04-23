package geo

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/vendors/local"

	mihomoresolver "github.com/metacubex/mihomo/component/resolver"
	D "github.com/miekg/dns"
)

func TestMacroRunCapturesInboundAndOutboundGeo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/":
			_, _ = fmt.Fprint(w, `{"success":true,"ip":"203.0.113.10","city":"Los Angeles","country":"United States","country_code":"US","connection":{"asn":64501,"org":"Exit Org","isp":"Exit ISP"}}`)
		case "/198.51.100.5":
			_, _ = fmt.Fprint(w, `{"success":true,"ip":"198.51.100.5","city":"Tokyo","country":"Japan","country_code":"JP","connection":{"asn":64500,"org":"Entry Org","isp":"Entry ISP"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	previousOutbound := outboundLookupURL
	previousLookupPattern := ipLookupURLPattern
	previousOutboundIPSB := outboundLookupURLIPSB
	previousLookupPatternIPSB := ipLookupURLPatternIPSB
	previousOutboundIPAPICo := outboundLookupURLIPAPICo
	previousLookupPatternIPAPICo := ipLookupURLPatternIPAPICo
	outboundLookupURL = server.URL + "/"
	ipLookupURLPattern = server.URL + "/%s"
	defer func() {
		outboundLookupURL = previousOutbound
		ipLookupURLPattern = previousLookupPattern
		outboundLookupURLIPSB = previousOutboundIPSB
		ipLookupURLPatternIPSB = previousLookupPatternIPSB
		outboundLookupURLIPAPICo = previousOutboundIPAPICo
		ipLookupURLPatternIPAPICo = previousLookupPatternIPAPICo
	}()

	task := &interfaces.Task{
		Config: interfaces.TaskConfig{
			TaskTimeoutMillis: 1000,
		},
	}
	vendor := (&local.Vendor{}).Build(interfaces.Node{
		Name:    "node-1",
		Payload: "198.51.100.5",
	})

	macro := &Macro{}
	if err := macro.Run(context.Background(), vendor, task); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if macro.Inbound.Address != "198.51.100.5" {
		t.Fatalf("expected inbound address 198.51.100.5, got %q", macro.Inbound.Address)
	}
	if macro.Inbound.IP != "198.51.100.5" {
		t.Fatalf("expected inbound IP 198.51.100.5, got %q", macro.Inbound.IP)
	}
	if macro.Inbound.ASN != 64500 {
		t.Fatalf("expected inbound ASN 64500, got %d", macro.Inbound.ASN)
	}
	if macro.Inbound.ASOrganization != "Entry Org" {
		t.Fatalf("expected inbound AS organization Entry Org, got %q", macro.Inbound.ASOrganization)
	}
	if macro.Inbound.City != "Tokyo" {
		t.Fatalf("expected inbound city Tokyo, got %q", macro.Inbound.City)
	}

	if macro.Outbound.IP != "203.0.113.10" {
		t.Fatalf("expected outbound IP 203.0.113.10, got %q", macro.Outbound.IP)
	}
	if macro.Outbound.ASN != 64501 {
		t.Fatalf("expected outbound ASN 64501, got %d", macro.Outbound.ASN)
	}
	if macro.Outbound.ASOrganization != "Exit Org" {
		t.Fatalf("expected outbound AS organization Exit Org, got %q", macro.Outbound.ASOrganization)
	}
	if macro.Outbound.City != "Los Angeles" {
		t.Fatalf("expected outbound city Los Angeles, got %q", macro.Outbound.City)
	}
}

func TestLookupIPInfoFallsBackToIPSBWhenPrimaryFails(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream blocked", http.StatusBadGateway)
	}))
	defer primary.Close()

	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/geoip/198.51.100.7":
			_, _ = fmt.Fprint(w, `{"ip":"198.51.100.7","country_code":"JP","country":"Japan","city":"Tokyo","asn":"64512","organization":"Fallback ISP"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer secondary.Close()

	previousOutbound := outboundLookupURL
	previousLookupPattern := ipLookupURLPattern
	previousOutboundIPSB := outboundLookupURLIPSB
	previousLookupPatternIPSB := ipLookupURLPatternIPSB
	previousOutboundIPAPICo := outboundLookupURLIPAPICo
	previousLookupPatternIPAPICo := ipLookupURLPatternIPAPICo
	outboundLookupURL = primary.URL + "/"
	ipLookupURLPattern = primary.URL + "/%s"
	outboundLookupURLIPSB = secondary.URL + "/geoip"
	ipLookupURLPatternIPSB = secondary.URL + "/geoip/%s"
	outboundLookupURLIPAPICo = primary.URL + "/json/"
	ipLookupURLPatternIPAPICo = primary.URL + "/%s/json/"
	defer func() {
		outboundLookupURL = previousOutbound
		ipLookupURLPattern = previousLookupPattern
		outboundLookupURLIPSB = previousOutboundIPSB
		ipLookupURLPatternIPSB = previousLookupPatternIPSB
		outboundLookupURLIPAPICo = previousOutboundIPAPICo
		ipLookupURLPatternIPAPICo = previousLookupPatternIPAPICo
	}()

	info, err := lookupIPInfo(context.Background(), nil, "198.51.100.7", 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("lookupIPInfo() error = %v", err)
	}
	if info.IP != "198.51.100.7" {
		t.Fatalf("expected fallback IP 198.51.100.7, got %q", info.IP)
	}
	if info.CountryCode != "JP" {
		t.Fatalf("expected fallback country JP, got %q", info.CountryCode)
	}
	if info.ASN != 64512 {
		t.Fatalf("expected fallback ASN 64512, got %d", info.ASN)
	}
	if info.ASOrganization != "Fallback ISP" {
		t.Fatalf("expected fallback organization Fallback ISP, got %q", info.ASOrganization)
	}
}

func TestLookupCurrentIPInfoFallsBackToIPAPICoWhenEarlierProvidersFail(t *testing.T) {
	blocked := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream blocked", http.StatusBadGateway)
	}))
	defer blocked.Close()

	tertiary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/json/":
			_, _ = fmt.Fprint(w, `{"ip":"203.0.113.9","city":"Los Angeles","country_name":"United States","country_code":"US","asn":"AS64513","org":"Fallback Exit Org"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer tertiary.Close()

	previousOutbound := outboundLookupURL
	previousLookupPattern := ipLookupURLPattern
	previousOutboundIPSB := outboundLookupURLIPSB
	previousLookupPatternIPSB := ipLookupURLPatternIPSB
	previousOutboundIPAPICo := outboundLookupURLIPAPICo
	previousLookupPatternIPAPICo := ipLookupURLPatternIPAPICo
	outboundLookupURL = blocked.URL + "/"
	ipLookupURLPattern = blocked.URL + "/%s"
	outboundLookupURLIPSB = blocked.URL + "/geoip"
	ipLookupURLPatternIPSB = blocked.URL + "/geoip/%s"
	outboundLookupURLIPAPICo = tertiary.URL + "/json/"
	ipLookupURLPatternIPAPICo = tertiary.URL + "/%s/json/"
	defer func() {
		outboundLookupURL = previousOutbound
		ipLookupURLPattern = previousLookupPattern
		outboundLookupURLIPSB = previousOutboundIPSB
		ipLookupURLPatternIPSB = previousLookupPatternIPSB
		outboundLookupURLIPAPICo = previousOutboundIPAPICo
		ipLookupURLPatternIPAPICo = previousLookupPatternIPAPICo
	}()

	info, err := lookupCurrentIPInfo(context.Background(), nil, 1500*time.Millisecond)
	if err != nil {
		t.Fatalf("lookupCurrentIPInfo() error = %v", err)
	}
	if info.IP != "203.0.113.9" {
		t.Fatalf("expected fallback IP 203.0.113.9, got %q", info.IP)
	}
	if info.CountryCode != "US" {
		t.Fatalf("expected fallback country US, got %q", info.CountryCode)
	}
	if info.ASN != 64513 {
		t.Fatalf("expected fallback ASN 64513, got %d", info.ASN)
	}
	if info.ASOrganization != "Fallback Exit Org" {
		t.Fatalf("expected fallback organization Fallback Exit Org, got %q", info.ASOrganization)
	}
}

func TestExtractHost(t *testing.T) {
	cases := map[string]string{
		"1.2.3.4:443":         "1.2.3.4",
		"example.com:8443":    "example.com",
		"[2001:db8::1]:443":   "2001:db8::1",
		"2001:db8::1":         "2001:db8::1",
		"direct":              "",
		"  example.com:443  ": "example.com",
	}

	for input, expected := range cases {
		if actual := extractHost(input); actual != expected {
			t.Fatalf("extractHost(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestResolveHostPrefersMihomoResolverForProxyHosts(t *testing.T) {
	original := mihomoresolver.ProxyServerHostResolver
	defer func() {
		mihomoresolver.ProxyServerHostResolver = original
	}()

	mihomoresolver.ProxyServerHostResolver = stubResolver{
		lookupIP: func(host string) ([]netip.Addr, error) {
			if host != "node.example.invalid" {
				t.Fatalf("unexpected host %q", host)
			}
			return []netip.Addr{netip.MustParseAddr("198.51.100.9")}, nil
		},
	}

	ip, err := resolveHost(t.Context(), "node.example.invalid")
	if err != nil {
		t.Fatalf("resolveHost() error = %v", err)
	}
	if ip != "198.51.100.9" {
		t.Fatalf("resolveHost() = %q, want %q", ip, "198.51.100.9")
	}
}

func TestMacroRunRetriesInboundAndOutboundIndependently(t *testing.T) {
	previousLookupInbound := lookupInboundFunc
	previousLookupOutbound := lookupOutboundFunc
	defer func() {
		lookupInboundFunc = previousLookupInbound
		lookupOutboundFunc = previousLookupOutbound
	}()

	inboundAttempts := 0
	outboundAttempts := 0
	lookupInboundFunc = func(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
		_ = ctx
		_ = proxy
		_ = timeout
		inboundAttempts++
		if inboundAttempts == 1 {
			return interfaces.GeoIPInfo{Address: "198.51.100.5"}, fmt.Errorf("temporary inbound failure")
		}
		return interfaces.GeoIPInfo{
			Address:     "198.51.100.5",
			IP:          "198.51.100.5",
			CountryCode: "JP",
		}, nil
	}
	lookupOutboundFunc = func(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
		_ = ctx
		_ = proxy
		_ = timeout
		outboundAttempts++
		if outboundAttempts == 1 {
			return interfaces.GeoIPInfo{}, fmt.Errorf("temporary outbound failure")
		}
		return interfaces.GeoIPInfo{
			IP:          "203.0.113.10",
			CountryCode: "US",
		}, nil
	}

	macro := &Macro{}
	err := macro.Run(context.Background(), (&local.Vendor{}).Build(interfaces.Node{Name: "node-1", Payload: "198.51.100.5"}), &interfaces.Task{
		Config: interfaces.TaskConfig{
			TaskRetry:         2,
			TaskTimeoutMillis: 1000,
		},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if inboundAttempts != 2 {
		t.Fatalf("expected 2 inbound attempts, got %d", inboundAttempts)
	}
	if outboundAttempts != 2 {
		t.Fatalf("expected 2 outbound attempts, got %d", outboundAttempts)
	}
	if macro.Inbound.CountryCode != "JP" {
		t.Fatalf("expected inbound country JP, got %q", macro.Inbound.CountryCode)
	}
	if macro.Outbound.CountryCode != "US" {
		t.Fatalf("expected outbound country US, got %q", macro.Outbound.CountryCode)
	}
}

func TestMacroRunReturnsSingleSideSuccessAfterOtherSideRetriesExhausted(t *testing.T) {
	previousLookupInbound := lookupInboundFunc
	previousLookupOutbound := lookupOutboundFunc
	defer func() {
		lookupInboundFunc = previousLookupInbound
		lookupOutboundFunc = previousLookupOutbound
	}()

	inboundAttempts := 0
	lookupInboundFunc = func(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
		_ = ctx
		_ = proxy
		_ = timeout
		inboundAttempts++
		return interfaces.GeoIPInfo{Address: "198.51.100.5"}, fmt.Errorf("inbound still failing")
	}
	lookupOutboundFunc = func(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.GeoIPInfo, error) {
		_ = ctx
		_ = proxy
		_ = timeout
		return interfaces.GeoIPInfo{
			IP:          "203.0.113.10",
			CountryCode: "US",
		}, nil
	}

	macro := &Macro{}
	err := macro.Run(context.Background(), (&local.Vendor{}).Build(interfaces.Node{Name: "node-1", Payload: "198.51.100.5"}), &interfaces.Task{
		Config: interfaces.TaskConfig{
			TaskRetry:         2,
			TaskTimeoutMillis: 1000,
		},
	})
	if err != nil {
		t.Fatalf("expected partial geo success to keep node alive, got %v", err)
	}
	if inboundAttempts != 2 {
		t.Fatalf("expected 2 inbound attempts, got %d", inboundAttempts)
	}
	if macro.Inbound.Error != "inbound still failing" {
		t.Fatalf("expected inbound error to be preserved, got %q", macro.Inbound.Error)
	}
	if macro.Outbound.CountryCode != "US" {
		t.Fatalf("expected outbound country US, got %q", macro.Outbound.CountryCode)
	}
}

type stubResolver struct {
	lookupIP func(host string) ([]netip.Addr, error)
}

func (s stubResolver) LookupIP(_ context.Context, host string) ([]netip.Addr, error) {
	return s.lookupIP(host)
}

func (s stubResolver) LookupIPv4(ctx context.Context, host string) ([]netip.Addr, error) {
	return s.LookupIP(ctx, host)
}

func (s stubResolver) LookupIPv6(_ context.Context, _ string) ([]netip.Addr, error) {
	return nil, mihomoresolver.ErrIPv6Disabled
}

func (s stubResolver) ResolveECH(_ context.Context, _ string) ([]byte, error) {
	return nil, mihomoresolver.ErrIPNotFound
}

func (s stubResolver) ExchangeContext(_ context.Context, _ *D.Msg) (*D.Msg, error) {
	return nil, mihomoresolver.ErrIPNotFound
}

func (s stubResolver) Invalid() bool {
	return true
}

func (s stubResolver) ClearCache() {}

func (s stubResolver) ResetConnection() {}
