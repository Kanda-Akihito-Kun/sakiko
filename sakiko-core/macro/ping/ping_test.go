package ping

import (
	"context"
	"fmt"
	"net"
	"testing"

	"sakiko.local/sakiko-core/interfaces"
)

type stubVendor struct{}

func (stubVendor) Type() interfaces.VendorType                  { return interfaces.VendorInvalid }
func (stubVendor) Status() interfaces.VendorStatus              { return interfaces.VStatusOperational }
func (stubVendor) Build(node interfaces.Node) interfaces.Vendor { return stubVendor{} }
func (stubVendor) DialTCP(ctx context.Context, rawURL string, network interfaces.RequestOptionsNetwork) (net.Conn, error) {
	_ = ctx
	_ = rawURL
	_ = network
	return nil, fmt.Errorf("not implemented")
}
func (stubVendor) DialUDP(ctx context.Context, rawURL string) (net.PacketConn, error) {
	_ = ctx
	_ = rawURL
	return nil, fmt.Errorf("not implemented")
}
func (stubVendor) ProxyInfo() interfaces.ProxyInfo { return interfaces.ProxyInfo{Name: "stub"} }

func TestPingRetriesEachSampleIndependently(t *testing.T) {
	originalTrace := pingViaTraceFunc
	originalNetcat := pingViaNetcatFunc
	defer func() {
		pingViaTraceFunc = originalTrace
		pingViaNetcatFunc = originalNetcat
	}()

	attempts := 0
	pingViaTraceFunc = func(ctx context.Context, proxy interfaces.Vendor, rawURL string) (uint16, uint16, error) {
		_ = ctx
		_ = proxy
		_ = rawURL
		attempts++
		switch attempts {
		case 1:
			return 0, 0, fmt.Errorf("temporary failure")
		case 2:
			return 120, 420, nil
		case 3:
			return 130, 430, nil
		default:
			return 0, 0, fmt.Errorf("unexpected extra attempt")
		}
	}

	rtt, delay := ping(stubVendor{}, "https://example.com", 2, 2, 1000)
	if rtt != 125 {
		t.Fatalf("expected RTT average 125, got %d", rtt)
	}
	if delay != 425 {
		t.Fatalf("expected delay average 425, got %d", delay)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestPingReturnsDelayWhenRTTIsUnavailable(t *testing.T) {
	originalTrace := pingViaTraceFunc
	originalNetcat := pingViaNetcatFunc
	defer func() {
		pingViaTraceFunc = originalTrace
		pingViaNetcatFunc = originalNetcat
	}()

	pingViaTraceFunc = func(ctx context.Context, proxy interfaces.Vendor, rawURL string) (uint16, uint16, error) {
		_ = ctx
		_ = proxy
		_ = rawURL
		return 0, 360, nil
	}

	rtt, delay := ping(stubVendor{}, "https://example.com", 1, 1, 1000)
	if rtt != 0 {
		t.Fatalf("expected RTT 0, got %d", rtt)
	}
	if delay != 360 {
		t.Fatalf("expected delay 360, got %d", delay)
	}
}
