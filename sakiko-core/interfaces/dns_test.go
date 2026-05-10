package interfaces

import "testing"

func TestDNSConfigNormalizeUsesDefaults(t *testing.T) {
	cfg := (DNSConfig{}).Normalize()

	if len(cfg.BootstrapServers) == 0 {
		t.Fatalf("expected default bootstrap servers")
	}
	if len(cfg.ResolverServers) == 0 {
		t.Fatalf("expected default resolver servers")
	}
}

func TestDNSConfigNormalizeTrimsAndDeduplicates(t *testing.T) {
	cfg := (DNSConfig{
		BootstrapServers: []string{" 1.1.1.1 ", "1.1.1.1", ""},
		ResolverServers:  []string{" https://dns.example/dns-query ", "https://dns.example/dns-query"},
	}).Normalize()

	if len(cfg.BootstrapServers) != 1 || cfg.BootstrapServers[0] != "1.1.1.1" {
		t.Fatalf("unexpected bootstrap servers: %#v", cfg.BootstrapServers)
	}
	if len(cfg.ResolverServers) != 1 || cfg.ResolverServers[0] != "https://dns.example/dns-query" {
		t.Fatalf("unexpected resolver servers: %#v", cfg.ResolverServers)
	}
}
