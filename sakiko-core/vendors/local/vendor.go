package local

import (
	"context"
	"fmt"
	"net"

	"sakiko.local/sakiko-core/interfaces"
)

type Vendor struct {
	node interfaces.Node
}

func (v *Vendor) Type() interfaces.VendorType {
	return interfaces.VendorLocal
}

func (v *Vendor) Status() interfaces.VendorStatus {
	return interfaces.VStatusOperational
}

func (v *Vendor) Build(node interfaces.Node) interfaces.Vendor {
	v.node = node
	return v
}

func (v *Vendor) DialTCP(ctx context.Context, rawURL string, network interfaces.RequestOptionsNetwork) (net.Conn, error) {
	host, port, err := urlToMetadata(rawURL)
	if err != nil {
		return nil, err
	}
	var d net.Dialer
	return d.DialContext(ctx, network.String(), fmt.Sprintf("%s:%d", host, port))
}

func (v *Vendor) DialUDP(ctx context.Context, rawURL string) (net.PacketConn, error) {
	host, port, err := urlToMetadata(rawURL)
	if err != nil {
		return nil, err
	}
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	return net.ListenUDP("udp", addr)
}

func (v *Vendor) ProxyInfo() interfaces.ProxyInfo {
	address := v.node.Payload
	if address == "" {
		address = "direct"
	}
	return interfaces.ProxyInfo{
		Name:    v.node.Name,
		Address: address,
		Type:    interfaces.ProxyUnknown,
	}
}
