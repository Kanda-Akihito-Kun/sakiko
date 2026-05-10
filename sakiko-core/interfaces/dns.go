package interfaces

import "strings"

var defaultDNSBootstrapServers = []string{
	"114.114.114.114",
	"223.5.5.5",
	"8.8.8.8",
	"1.0.0.1",
}

var defaultDNSResolverServers = []string{
	"https://doh.pub/dns-query",
	"https://dns.alidns.com/dns-query",
	"tls://223.5.5.5:853",
}

type DNSConfig struct {
	BootstrapServers []string `json:"bootstrapServers"`
	ResolverServers  []string `json:"resolverServers"`
}

func DefaultDNSConfig() DNSConfig {
	return DNSConfig{
		BootstrapServers: append([]string{}, defaultDNSBootstrapServers...),
		ResolverServers:  append([]string{}, defaultDNSResolverServers...),
	}
}

func (c DNSConfig) Normalize() DNSConfig {
	bootstrap := normalizeNameServers(c.BootstrapServers)
	resolvers := normalizeNameServers(c.ResolverServers)

	if len(bootstrap) == 0 {
		bootstrap = append([]string{}, defaultDNSBootstrapServers...)
	}
	if len(resolvers) == 0 {
		resolvers = append([]string{}, defaultDNSResolverServers...)
	}

	return DNSConfig{
		BootstrapServers: bootstrap,
		ResolverServers:  resolvers,
	}
}

func normalizeNameServers(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
