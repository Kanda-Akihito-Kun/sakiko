package mihomo

import (
	"context"
	"net/netip"
	"testing"

	"sakiko.local/sakiko-core/interfaces"

	"github.com/metacubex/mihomo/component/resolver"
	D "github.com/miekg/dns"
)

func TestEnsureRuntimeInitializesResolvers(t *testing.T) {
	originalDefault := resolver.DefaultResolver
	originalProxy := resolver.ProxyServerHostResolver
	originalDirect := resolver.DirectHostResolver
	originalConfig := runtimeDNSConfig
	defer func() {
		resolver.DefaultResolver = originalDefault
		resolver.ProxyServerHostResolver = originalProxy
		resolver.DirectHostResolver = originalDirect
		runtimeDNSConfig = originalConfig
	}()

	resolver.DefaultResolver = nil
	resolver.ProxyServerHostResolver = nil
	resolver.DirectHostResolver = nil

	ensureRuntime()

	if resolver.DefaultResolver == nil {
		t.Fatalf("expected default resolver to be initialized")
	}
	if resolver.ProxyServerHostResolver == nil {
		t.Fatalf("expected proxy server host resolver to be initialized")
	}
	if resolver.DirectHostResolver == nil {
		t.Fatalf("expected direct host resolver to be initialized")
	}
}

func TestConfigureDNSConfigAppliesNormalizedConfig(t *testing.T) {
	originalConfig := runtimeDNSConfig
	defer func() {
		runtimeDNSConfig = originalConfig
	}()

	err := ConfigureDNSConfig(interfaces.DNSConfig{
		BootstrapServers: []string{" 223.5.5.5 ", "223.5.5.5"},
		ResolverServers:  []string{" https://dns.alidns.com/dns-query "},
	})
	if err != nil {
		t.Fatalf("ConfigureDNSConfig() error = %v", err)
	}

	current := CurrentDNSConfig()
	if len(current.BootstrapServers) != 1 || current.BootstrapServers[0] != "223.5.5.5" {
		t.Fatalf("unexpected bootstrap servers: %#v", current.BootstrapServers)
	}
	if len(current.ResolverServers) != 1 || current.ResolverServers[0] != "https://dns.alidns.com/dns-query" {
		t.Fatalf("unexpected resolver servers: %#v", current.ResolverServers)
	}
}

func TestResolveHostIPsUsesProxyResolverAndPrefersIPv4(t *testing.T) {
	originalProxy := resolver.ProxyServerHostResolver
	defer func() {
		resolver.ProxyServerHostResolver = originalProxy
	}()

	resolver.ProxyServerHostResolver = stubResolver{
		lookupIP: func(host string) ([]netip.Addr, error) {
			if host != "node.example.invalid" {
				t.Fatalf("unexpected host %q", host)
			}
			return []netip.Addr{
				netip.MustParseAddr("2001:db8::1"),
				netip.MustParseAddr("198.51.100.9"),
				netip.MustParseAddr("198.51.100.9"),
			}, nil
		},
	}

	ips, err := ResolveHostIPs(context.Background(), "node.example.invalid")
	if err != nil {
		t.Fatalf("ResolveHostIPs() error = %v", err)
	}
	if len(ips) != 2 {
		t.Fatalf("expected 2 deduped results, got %d", len(ips))
	}
	if ips[0] != "198.51.100.9" {
		t.Fatalf("expected IPv4 result first, got %q", ips[0])
	}
	if ips[1] != "2001:db8::1" {
		t.Fatalf("expected IPv6 result second, got %q", ips[1])
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
	return nil, resolver.ErrIPv6Disabled
}

func (s stubResolver) ResolveECH(_ context.Context, _ string) ([]byte, error) {
	return nil, resolver.ErrIPNotFound
}

func (s stubResolver) ExchangeContext(_ context.Context, _ *D.Msg) (*D.Msg, error) {
	return nil, resolver.ErrIPNotFound
}

func (s stubResolver) Invalid() bool {
	return true
}

func (s stubResolver) ClearCache() {}

func (s stubResolver) ResetConnection() {}
