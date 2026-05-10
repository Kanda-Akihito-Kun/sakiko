package cluster

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	udpmacro "sakiko.local/sakiko-core/macro/udp"
	"sakiko.local/sakiko-core/netenv"
	localvendor "sakiko.local/sakiko-core/vendors/local"
)

const (
	natTypeUnknown = "unknown"
	natTypeNAT1    = "nat1"
	natTypeNAT2    = "nat2"
	natTypeNAT3    = "nat3"
	natTypeNAT4    = "nat4"
)

type defaultMasterEligibilityProber struct{}

var (
	hasDirectPublicIPFunc            = hasDirectPublicIP
	probeLocalUDPNATFunc             = probeLocalUDPNAT
	probePublicListenerReachableFunc = probePublicListenerReachable
)

func (defaultMasterEligibilityProber) ProbeMasterEligibility(ctx context.Context) interfaces.MasterEligibility {
	info := netenv.Probe(ctx)
	return evaluateMasterEligibility(ctx, info, time.Now().UTC())
}

func evaluateMasterEligibility(ctx context.Context, info interfaces.BackendInfo, checkedAt time.Time) interfaces.MasterEligibility {
	eligibility := interfaces.MasterEligibility{
		PublicIP:  strings.TrimSpace(info.IP),
		NATType:   natTypeUnknown,
		CheckedAt: checkedAt.Format(time.RFC3339),
	}

	if errText := strings.TrimSpace(info.Error); errText != "" {
		eligibility.Error = errText
		return eligibility
	}

	eligibility.HasPublicIP = isPublicIPAddress(eligibility.PublicIP)
	if !eligibility.HasPublicIP {
		eligibility.Error = "public IP is required before enabling master mode"
		return eligibility
	}

	if hasDirectPublicIPFunc(eligibility.PublicIP) {
		markNAT1(&eligibility)
		return eligibility
	}

	natInfo := probeLocalUDPNATFunc(ctx)
	applyLocalNATEligibility(ctx, &eligibility, natInfo)
	return eligibility
}

func applyLocalNATEligibility(ctx context.Context, eligibility *interfaces.MasterEligibility, natInfo interfaces.UDPNATInfo) {
	if eligibility == nil {
		return
	}

	switch natInfo.Type {
	case interfaces.UDPNATTypeOpen:
		markNAT1(eligibility)
		return
	case interfaces.UDPNATTypeFullCone:
		if canServeAsPublicNAT1(ctx, eligibility.PublicIP, natInfo) {
			markNAT1(eligibility)
			return
		}
		eligibility.NATType = natTypeNAT2
	case interfaces.UDPNATTypeRestrictedCone, interfaces.UDPNATTypeRestrictedPort:
		eligibility.NATType = natTypeNAT3
	case interfaces.UDPNATTypeSymmetric, interfaces.UDPNATTypeUDPFirewall, interfaces.UDPNATTypeBlocked:
		eligibility.NATType = natTypeNAT4
	default:
		eligibility.NATType = natTypeUnknown
	}

	if errText := strings.TrimSpace(natInfo.Error); errText != "" {
		eligibility.Error = errText
		return
	}

	eligibility.Error = "a direct public IP or nat1 network is required before enabling master mode"
}

func canServeAsPublicNAT1(ctx context.Context, publicIP string, natInfo interfaces.UDPNATInfo) bool {
	if natInfo.Type == interfaces.UDPNATTypeOpen {
		return true
	}
	if natInfo.Type != interfaces.UDPNATTypeFullCone {
		return false
	}
	if !sameIPAddress(publicIP, natInfo.PublicIP) {
		return false
	}

	reachable, err := probePublicListenerReachableFunc(ctx, publicIP)
	return err == nil && reachable
}

func markNAT1(eligibility *interfaces.MasterEligibility) {
	if eligibility == nil {
		return
	}
	eligibility.NATType = natTypeNAT1
	eligibility.IsNAT1 = true
	eligibility.Reachable = true
	eligibility.Eligible = true
	eligibility.Error = ""
}

func hasDirectPublicIP(publicIP string) bool {
	target, err := netip.ParseAddr(strings.TrimSpace(publicIP))
	if err != nil || !target.IsValid() {
		return false
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}

	for _, addr := range addrs {
		prefix, err := netip.ParsePrefix(strings.TrimSpace(addr.String()))
		if err == nil && prefix.Addr() == target {
			return true
		}
	}
	return false
}

func probeLocalUDPNAT(ctx context.Context) interfaces.UDPNATInfo {
	if ctx == nil {
		ctx = context.Background()
	}

	vendor := (&localvendor.Vendor{}).Build(interfaces.Node{Name: "master-eligibility", Payload: "direct"})
	macro := &udpmacro.Macro{}
	_ = macro.Run(ctx, vendor, nil)
	return macro.Info
}

func probePublicListenerReachable(ctx context.Context, publicIP string) (bool, error) {
	host := strings.TrimSpace(publicIP)
	if !isPublicIPAddress(host) {
		return false, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false, err
	}

	network := "tcp4"
	listenAddr := "0.0.0.0:0"
	if addr.Is6() {
		network = "tcp6"
		listenAddr = "[::]:0"
	}

	listener, err := net.Listen(network, listenAddr)
	if err != nil {
		return false, err
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	accepted := make(chan struct{}, 1)
	acceptErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				acceptErr <- err
			}
			return
		}
		_ = conn.Close()
		accepted <- struct{}{}
	}()

	timeout := 2 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		if remaining := time.Until(deadline); remaining > 0 && remaining < timeout {
			timeout = remaining
		}
	}
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(dialCtx, network, net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return false, err
	}
	_ = conn.Close()

	select {
	case <-accepted:
		return true, nil
	case err := <-acceptErr:
		return false, err
	case <-dialCtx.Done():
		return false, dialCtx.Err()
	}
}

func sameIPAddress(left string, right string) bool {
	leftAddr, err := netip.ParseAddr(strings.TrimSpace(left))
	if err != nil {
		return false
	}
	rightAddr, err := netip.ParseAddr(strings.TrimSpace(right))
	if err != nil {
		return false
	}
	return leftAddr == rightAddr
}

func isPublicIPAddress(raw string) bool {
	addr, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	if !addr.IsValid() {
		return false
	}
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() {
		return false
	}
	return true
}
