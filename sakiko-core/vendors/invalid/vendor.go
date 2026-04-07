package invalid

import (
	"context"
	"fmt"
	"net"

	"sakiko.local/sakiko-core/interfaces"
)

type Vendor struct{}

func (v *Vendor) Type() interfaces.VendorType {
	return interfaces.VendorInvalid
}

func (v *Vendor) Status() interfaces.VendorStatus {
	return interfaces.VStatusNotReady
}

func (v *Vendor) Build(node interfaces.Node) interfaces.Vendor {
	_ = node
	return v
}

func (v *Vendor) DialTCP(ctx context.Context, rawURL string, network interfaces.RequestOptionsNetwork) (net.Conn, error) {
	_ = ctx
	_ = rawURL
	_ = network
	return nil, fmt.Errorf("invalid vendor")
}

func (v *Vendor) DialUDP(ctx context.Context, rawURL string) (net.PacketConn, error) {
	_ = ctx
	_ = rawURL
	return nil, fmt.Errorf("invalid vendor")
}

func (v *Vendor) ProxyInfo() interfaces.ProxyInfo {
	return interfaces.ProxyInfo{Type: interfaces.ProxyUnknown}
}
