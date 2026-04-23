package mihomo

import (
	"testing"

	"github.com/metacubex/mihomo/component/resolver"
)

func TestEnsureRuntimeInitializesResolvers(t *testing.T) {
	originalDefault := resolver.DefaultResolver
	originalProxy := resolver.ProxyServerHostResolver
	originalDirect := resolver.DirectHostResolver
	defer func() {
		resolver.DefaultResolver = originalDefault
		resolver.ProxyServerHostResolver = originalProxy
		resolver.DirectHostResolver = originalDirect
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
