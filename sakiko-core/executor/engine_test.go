package executor

import (
	"context"
	"net"
	"reflect"
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func TestMacroExecutionOrderPrioritizesHeavyAndAuxiliaryMacrosLast(t *testing.T) {
	input := []interfaces.MacroType{
		interfaces.MacroMedia,
		interfaces.MacroSpeed,
		interfaces.MacroPing,
		interfaces.MacroGeo,
	}

	got := macroExecutionOrder(input)
	want := []interfaces.MacroType{
		interfaces.MacroPing,
		interfaces.MacroGeo,
		interfaces.MacroSpeed,
		interfaces.MacroMedia,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("macroExecutionOrder() = %v, want %v", got, want)
	}
}

func TestRunMacrosAppliesIntervalBetweenMacroExecutions(t *testing.T) {
	originalFindMacro := findMacroFunc
	originalWaitMacroInterval := waitMacroIntervalFunc
	defer func() {
		findMacroFunc = originalFindMacro
		waitMacroIntervalFunc = originalWaitMacroInterval
	}()

	var executed []interfaces.MacroType
	var waits []time.Duration
	findMacroFunc = func(macroType interfaces.MacroType) interfaces.Macro {
		return stubMacro{
			macroType: macroType,
			run: func() {
				executed = append(executed, macroType)
			},
		}
	}
	waitMacroIntervalFunc = func(ctx context.Context, interval time.Duration) error {
		_ = ctx
		waits = append(waits, interval)
		return nil
	}

	_, err := runMacros(
		context.Background(),
		stubVendor{proxyInfo: interfaces.ProxyInfo{Name: "node-1", Address: "example.com:443"}},
		&interfaces.Task{},
		[]interfaces.MatrixEntry{
			{Type: interfaces.MatrixRTTPing},
			{Type: interfaces.MatrixInboundGeoIP},
			{Type: interfaces.MatrixAverageSpeed},
		},
		[]interfaces.MacroType{interfaces.MacroSpeed, interfaces.MacroPing, interfaces.MacroGeo},
		nil,
	)
	if err != nil {
		t.Fatalf("runMacros() error = %v", err)
	}

	wantExecuted := []interfaces.MacroType{
		interfaces.MacroPing,
		interfaces.MacroGeo,
		interfaces.MacroSpeed,
	}
	if !reflect.DeepEqual(executed, wantExecuted) {
		t.Fatalf("executed macros = %v, want %v", executed, wantExecuted)
	}

	wantWaits := []time.Duration{defaultMacroInterval, defaultMacroInterval}
	if !reflect.DeepEqual(waits, wantWaits) {
		t.Fatalf("wait intervals = %v, want %v", waits, wantWaits)
	}
}

type stubMacro struct {
	macroType interfaces.MacroType
	run       func()
}

func (m stubMacro) Type() interfaces.MacroType {
	return m.macroType
}

func (m stubMacro) Run(ctx context.Context, proxy interfaces.Vendor, task *interfaces.Task) error {
	_, _, _ = ctx, proxy, task
	if m.run != nil {
		m.run()
	}
	return nil
}

type stubVendor struct {
	proxyInfo interfaces.ProxyInfo
}

func (v stubVendor) Type() interfaces.VendorType {
	return interfaces.VendorInvalid
}

func (v stubVendor) Status() interfaces.VendorStatus {
	return interfaces.VStatusOperational
}

func (v stubVendor) Build(node interfaces.Node) interfaces.Vendor {
	_ = node
	return v
}

func (v stubVendor) DialTCP(ctx context.Context, rawURL string, network interfaces.RequestOptionsNetwork) (net.Conn, error) {
	_, _, _ = ctx, rawURL, network
	return nil, nil
}

func (v stubVendor) DialUDP(ctx context.Context, rawURL string) (net.PacketConn, error) {
	_, _ = ctx, rawURL
	return nil, nil
}

func (v stubVendor) ProxyInfo() interfaces.ProxyInfo {
	return v.proxyInfo
}
