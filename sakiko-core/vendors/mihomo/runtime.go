package mihomo

import (
	"sync"

	"github.com/metacubex/mihomo/component/resolver"
	_ "github.com/metacubex/mihomo/config"
	MihomoDNS "github.com/metacubex/mihomo/dns"
)

var runtimeLock sync.Mutex

var (
	defaultNameServers = []string{
		"114.114.114.114",
		"223.5.5.5",
		"8.8.8.8",
		"1.0.0.1",
	}
	runtimeNameServers = []string{
		"https://doh.pub/dns-query",
		"https://dns.alidns.com/dns-query",
		"tls://223.5.5.5:853",
	}
)

func ensureRuntime() {
	runtimeLock.Lock()
	defer runtimeLock.Unlock()

	if resolver.DefaultResolver != nil && resolver.ProxyServerHostResolver != nil && resolver.DirectHostResolver != nil {
		return
	}

	defaultResolvers, err := MihomoDNS.ParseNameServer(defaultNameServers)
	if err != nil || len(defaultResolvers) == 0 {
		return
	}
	mainResolvers, err := MihomoDNS.ParseNameServer(runtimeNameServers)
	if err != nil || len(mainResolvers) == 0 {
		return
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
}
