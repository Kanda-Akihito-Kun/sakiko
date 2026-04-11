package mihomo

import (
	"context"
	"fmt"
	"net"
	"strings"

	"sakiko.local/sakiko-core/interfaces"

	"github.com/metacubex/mihomo/adapter"
	C "github.com/metacubex/mihomo/constant"
	"go.yaml.in/yaml/v3"
)

type Vendor struct {
	node  interfaces.Node
	proxy C.Proxy
}

func (v *Vendor) Type() interfaces.VendorType {
	return interfaces.VendorMihomo
}

func (v *Vendor) Status() interfaces.VendorStatus {
	if v == nil || v.proxy == nil {
		return interfaces.VStatusNotReady
	}
	return interfaces.VStatusOperational
}

func (v *Vendor) Build(node interfaces.Node) interfaces.Vendor {
	v.node = node
	raw := strings.TrimSpace(node.Payload)
	if raw == "" {
		return v
	}

	ensureRuntime()

	payload := map[string]any{}
	if err := yaml.Unmarshal([]byte(raw), &payload); err != nil {
		return v
	}

	proxy, err := adapter.ParseProxy(payload)
	if err != nil {
		return v
	}

	v.proxy = proxy
	return v
}

func (v *Vendor) DialTCP(ctx context.Context, rawURL string, _ interfaces.RequestOptionsNetwork) (net.Conn, error) {
	if v == nil || v.proxy == nil {
		return nil, fmt.Errorf("mihomo proxy is not ready")
	}

	metadata, err := urlToMetadata(rawURL, C.TCP)
	if err != nil {
		return nil, err
	}

	return v.proxy.DialContext(ctx, metadata)
}

func (v *Vendor) DialUDP(ctx context.Context, rawURL string) (net.PacketConn, error) {
	if v == nil || v.proxy == nil {
		return nil, fmt.Errorf("mihomo proxy is not ready")
	}

	metadata, err := urlToMetadata(rawURL, C.UDP)
	if err != nil {
		return nil, err
	}

	return v.proxy.ListenPacketContext(ctx, metadata)
}

func (v *Vendor) ProxyInfo() interfaces.ProxyInfo {
	if v == nil || v.proxy == nil {
		return interfaces.ProxyInfo{
			Name:    v.node.Name,
			Address: strings.TrimSpace(v.node.Payload),
			Type:    interfaces.ProxyUnknown,
		}
	}

	return interfaces.ProxyInfo{
		Name:    v.proxy.Name(),
		Address: v.proxy.Addr(),
		Type:    interfaces.ParseProxyType(strings.ToLower(v.proxy.Type().String())),
	}
}
