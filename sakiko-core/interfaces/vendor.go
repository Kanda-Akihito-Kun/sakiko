package interfaces

import (
	"context"
	"net"
)

type VendorType string

const (
	VendorMihomo  VendorType = "mihomo"
	VendorLocal   VendorType = "local"
	VendorInvalid VendorType = "invalid"
)

type VendorStatus uint

const (
	VStatusOperational VendorStatus = iota
	VStatusNotReady
)

type ProxyType string

const (
	ProxyUnknown     ProxyType = "unknown"
	ProxyShadowsocks ProxyType = "shadowsocks"
	ProxySSR         ProxyType = "ssr"
	ProxySocks5      ProxyType = "socks5"
	ProxyHTTP        ProxyType = "http"
	ProxyVMess       ProxyType = "vmess"
	ProxyTrojan      ProxyType = "trojan"
	ProxyVLESS       ProxyType = "vless"
	ProxyHysteria    ProxyType = "hysteria"
	ProxyHysteria2   ProxyType = "hysteria2"
	ProxyTuic        ProxyType = "tuic"
	ProxyAnyTLS      ProxyType = "anytls"
)

func ParseProxyType(raw string) ProxyType {
	switch ProxyType(raw) {
	case ProxyShadowsocks, ProxySSR, ProxySocks5, ProxyHTTP, ProxyVMess, ProxyTrojan, ProxyVLESS, ProxyHysteria, ProxyHysteria2, ProxyTuic, ProxyAnyTLS:
		return ProxyType(raw)
	default:
		return ProxyUnknown
	}
}

type ProxyInfo struct {
	Name    string    `json:"name"`
	Address string    `json:"address"`
	Type    ProxyType `json:"type"`
}

type Vendor interface {
	Type() VendorType
	Status() VendorStatus
	Build(node Node) Vendor
	DialTCP(ctx context.Context, rawURL string, network RequestOptionsNetwork) (net.Conn, error)
	DialUDP(ctx context.Context, rawURL string) (net.PacketConn, error)
	ProxyInfo() ProxyInfo
}
