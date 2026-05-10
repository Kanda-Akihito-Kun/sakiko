package mihomo

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"sakiko.local/sakiko-core/interfaces"

	"github.com/metacubex/mihomo/component/resolver"
	_ "github.com/metacubex/mihomo/config"
	MihomoDNS "github.com/metacubex/mihomo/dns"
)

var runtimeLock sync.Mutex

var runtimeDNSConfig = interfaces.DefaultDNSConfig()

func ensureRuntime() {
	runtimeLock.Lock()
	defer runtimeLock.Unlock()

	if resolver.DefaultResolver != nil && resolver.ProxyServerHostResolver != nil && resolver.DirectHostResolver != nil {
		return
	}
	_ = applyResolverConfigLocked(runtimeDNSConfig)
}

func ConfigureDNSConfig(cfg interfaces.DNSConfig) error {
	runtimeLock.Lock()
	defer runtimeLock.Unlock()

	cfg = cfg.Normalize()
	if err := applyResolverConfigLocked(cfg); err != nil {
		return err
	}
	runtimeDNSConfig = cfg
	return nil
}

func CurrentDNSConfig() interfaces.DNSConfig {
	runtimeLock.Lock()
	defer runtimeLock.Unlock()

	return runtimeDNSConfig.Normalize()
}

func applyResolverConfigLocked(cfg interfaces.DNSConfig) error {
	cfg = cfg.Normalize()

	defaultResolvers, err := MihomoDNS.ParseNameServer(cfg.BootstrapServers)
	if err != nil || len(defaultResolvers) == 0 {
		if err == nil {
			err = fmt.Errorf("bootstrap resolvers are empty")
		}
		return err
	}
	mainResolvers, err := MihomoDNS.ParseNameServer(cfg.ResolverServers)
	if err != nil || len(mainResolvers) == 0 {
		if err == nil {
			err = fmt.Errorf("resolver servers are empty")
		}
		return err
	}

	resolvers := MihomoDNS.NewResolver(MihomoDNS.Config{
		Main:         mainResolvers,
		Default:      defaultResolvers,
		ProxyServer:  mainResolvers,
		DirectServer: mainResolvers,
		IPv6:         !resolver.DisableIPv6,
		IPv6Timeout:  100,
	})

	resolver.DefaultResolver = resolvers
	if resolvers.ProxyResolver != nil && resolvers.ProxyResolver.Invalid() {
		resolver.ProxyServerHostResolver = resolvers.ProxyResolver
	} else {
		resolver.ProxyServerHostResolver = resolvers.Resolver
	}
	if resolvers.DirectResolver != nil && resolvers.DirectResolver.Invalid() {
		resolver.DirectHostResolver = resolvers.DirectResolver
	} else {
		resolver.DirectHostResolver = resolvers.Resolver
	}
	return nil
}

func ResolveHost(ctx context.Context, host string) (string, error) {
	ips, err := ResolveHostIPs(ctx, host)
	if err != nil {
		return "", err
	}
	return ips[0], nil
}

func ResolveHostIPs(ctx context.Context, host string) ([]string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	trimmed := strings.Trim(host, "[]")
	if ip := net.ParseIP(trimmed); ip != nil {
		return []string{ip.String()}, nil
	}

	targetResolver := resolver.ProxyServerHostResolver
	if targetResolver == nil {
		ensureRuntime()
		targetResolver = resolver.ProxyServerHostResolver
	}
	if targetResolver == nil {
		return nil, fmt.Errorf("mihomo proxy resolver is not ready")
	}

	addrs, err := resolver.LookupIPWithResolver(ctx, trimmed, targetResolver)
	if err != nil {
		return nil, err
	}

	ipv4 := make([]string, 0, len(addrs))
	ipv6 := make([]string, 0, len(addrs))
	seen := make(map[string]struct{}, len(addrs))
	for _, addr := range addrs {
		value := strings.TrimSpace(addr.String())
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		if addr.Is4() {
			ipv4 = append(ipv4, value)
			continue
		}
		ipv6 = append(ipv6, value)
	}

	out := append(ipv4, ipv6...)
	if len(out) == 0 {
		return nil, fmt.Errorf("no ip address resolved for %s", trimmed)
	}
	return out, nil
}
