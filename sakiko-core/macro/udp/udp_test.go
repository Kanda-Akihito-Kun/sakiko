package udp

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func TestMacroRunStoresDetectorResult(t *testing.T) {
	original := detectUDPNATTypeFunc
	defer func() {
		detectUDPNATTypeFunc = original
	}()

	expected := interfaces.UDPNATInfo{
		Type:         interfaces.UDPNATTypeFullCone,
		InternalIP:   "10.0.0.2",
		InternalPort: 40000,
		PublicIP:     "198.51.100.8",
		PublicPort:   50000,
	}
	detectUDPNATTypeFunc = func(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.UDPNATInfo, error) {
		_ = ctx
		_ = proxy
		_ = timeout
		return expected, nil
	}

	var macro Macro
	if err := macro.Run(context.Background(), stubVendor{}, &interfaces.Task{}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if macro.Info != expected {
		t.Fatalf("expected %+v, got %+v", expected, macro.Info)
	}
}

func TestMacroRunStoresDetectorErrorWithoutFailingNode(t *testing.T) {
	original := detectUDPNATTypeFunc
	defer func() {
		detectUDPNATTypeFunc = original
	}()

	detectUDPNATTypeFunc = func(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.UDPNATInfo, error) {
		_ = ctx
		_ = proxy
		_ = timeout
		return interfaces.UDPNATInfo{}, errors.New("stun timeout")
	}

	var macro Macro
	if err := macro.Run(context.Background(), stubVendor{}, &interfaces.Task{}); err != nil {
		t.Fatalf("expected UDP NAT detector error to be stored as payload only, got %v", err)
	}
	if macro.Info.Error != "stun timeout" {
		t.Fatalf("expected stored probe error, got %#v", macro.Info.Error)
	}
}

func TestClassifyTopologyReturnsSymmetric(t *testing.T) {
	original := bindingTestFunc
	defer func() {
		bindingTestFunc = original
	}()

	call := 0
	bindingTestFunc = func(ctx context.Context, conn net.PacketConn, server stunServer, deadline time.Time, changeIP bool, changePort bool) (*stunResponse, error) {
		_ = ctx
		_ = conn
		_ = server
		_ = deadline
		call++
		switch call {
		case 1:
			if !changeIP || !changePort {
				t.Fatalf("expected test 2 flags on first follow-up probe")
			}
			return nil, errSTUNTimeout
		case 2:
			if changeIP || changePort {
				t.Fatalf("expected plain binding request on changed address probe")
			}
			return &stunResponse{
				External: stunEndpoint{IP: "198.51.100.99", Port: 55000},
			}, nil
		default:
			t.Fatalf("unexpected extra probe %d", call)
			return nil, nil
		}
	}

	kind, err := classifyTopology(context.Background(), nil, stunServer{Host: "stun.example.com", Port: 3478}, time.Now().Add(time.Second), stunEndpoint{IP: "10.0.0.2", Port: 40000}, stunResponse{
		External: stunEndpoint{IP: "198.51.100.8", Port: 50000},
		Changed:  stunEndpoint{IP: "203.0.113.8", Port: 3479},
	})
	if err != nil {
		t.Fatalf("classifyTopology() error = %v", err)
	}
	if kind != interfaces.UDPNATTypeSymmetric {
		t.Fatalf("expected symmetric NAT, got %q", kind)
	}
}

func TestParseBindingResponsePrefersXORMappedAddress(t *testing.T) {
	var token [16]byte
	binary.BigEndian.PutUint32(token[:4], stunMagicCookie)
	copy(token[4:], []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})

	raw := buildBindingResponseForTest(token, stunEndpoint{IP: "198.51.100.8", Port: 50000}, stunEndpoint{IP: "203.0.113.8", Port: 3478})

	resp, err := parseBindingResponse(raw, token)
	if err != nil {
		t.Fatalf("parseBindingResponse() error = %v", err)
	}
	if resp.External.IP != "198.51.100.8" || resp.External.Port != 50000 {
		t.Fatalf("unexpected external endpoint %+v", resp.External)
	}
	if resp.Changed.IP != "203.0.113.8" || resp.Changed.Port != 3478 {
		t.Fatalf("unexpected changed endpoint %+v", resp.Changed)
	}
}

func buildBindingResponseForTest(token [16]byte, external stunEndpoint, changed stunEndpoint) []byte {
	attrs := make([]byte, 0, 32)
	attrs = append(attrs, encodeAddressAttribute(stunAttrXORMappedAddress, external, token)...)
	attrs = append(attrs, encodeAddressAttribute(stunAttrChangedAddress, changed, token)...)

	raw := make([]byte, 20+len(attrs))
	binary.BigEndian.PutUint16(raw[0:2], stunBindingResponse)
	binary.BigEndian.PutUint16(raw[2:4], uint16(len(attrs)))
	copy(raw[4:20], token[:])
	copy(raw[20:], attrs)
	return raw
}

func encodeAddressAttribute(attrType uint16, endpoint stunEndpoint, token [16]byte) []byte {
	ip := net.ParseIP(endpoint.IP).To4()
	value := make([]byte, 8)
	value[1] = 0x01
	port := uint16(endpoint.Port)
	if attrType == stunAttrXORMappedAddress {
		binary.BigEndian.PutUint16(value[2:4], port^uint16(stunMagicCookie>>16))
		for i := 0; i < 4; i++ {
			value[4+i] = ip[i] ^ token[i]
		}
	} else {
		binary.BigEndian.PutUint16(value[2:4], port)
		copy(value[4:8], ip)
	}

	attr := make([]byte, 4+len(value))
	binary.BigEndian.PutUint16(attr[0:2], attrType)
	binary.BigEndian.PutUint16(attr[2:4], uint16(len(value)))
	copy(attr[4:], value)
	return attr
}

type stubVendor struct{}

func (stubVendor) Type() interfaces.VendorType                  { return interfaces.VendorInvalid }
func (stubVendor) Status() interfaces.VendorStatus              { return interfaces.VStatusOperational }
func (stubVendor) Build(node interfaces.Node) interfaces.Vendor { return stubVendor{} }
func (stubVendor) DialTCP(ctx context.Context, rawURL string, network interfaces.RequestOptionsNetwork) (net.Conn, error) {
	_ = ctx
	_ = rawURL
	_ = network
	return nil, nil
}
func (stubVendor) DialUDP(ctx context.Context, rawURL string) (net.PacketConn, error) {
	_ = ctx
	_ = rawURL
	return nil, nil
}
func (stubVendor) ProxyInfo() interfaces.ProxyInfo { return interfaces.ProxyInfo{Name: "stub"} }
