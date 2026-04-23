package udp

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

const (
	stunBindingRequest  uint16 = 0x0001
	stunBindingResponse uint16 = 0x0101

	stunAttrMappedAddress    uint16 = 0x0001
	stunAttrChangeRequest    uint16 = 0x0003
	stunAttrSourceAddress    uint16 = 0x0004
	stunAttrChangedAddress   uint16 = 0x0005
	stunAttrXORMappedAddress uint16 = 0x0020
	stunAttrResponseOrigin   uint16 = 0x802b
	stunAttrOtherAddress     uint16 = 0x802c

	stunMagicCookie uint32 = 0x2112A442

	minProbeTimeout    = 5 * time.Second
	requestRetryWindow = 500 * time.Millisecond
	maxRequestRetries  = 3
)

var (
	stunServers = []stunServer{
		{Host: "stun.ekiga.net", Port: 3478},
		{Host: "stun.ideasip.com", Port: 3478},
		{Host: "stun.voiparound.com", Port: 3478},
		{Host: "stun.voipbuster.com", Port: 3478},
		{Host: "stun.voipstunt.com", Port: 3478},
		{Host: "stun.voxgratia.org", Port: 3478},
	}
	detectUDPNATTypeFunc = detectUDPNATType
	bindingTestFunc      = bindingTest
)

type Macro struct {
	Info interfaces.UDPNATInfo
}

type stunServer struct {
	Host string
	Port int
}

type stunEndpoint struct {
	IP   string
	Port int
}

type stunResponse struct {
	External stunEndpoint
	Source   stunEndpoint
	Changed  stunEndpoint
}

func (m *Macro) Type() interfaces.MacroType {
	return interfaces.MacroUDP
}

func (m *Macro) Run(ctx context.Context, proxy interfaces.Vendor, task *interfaces.Task) error {
	timeout := resolveProbeTimeout(task)
	info, err := detectUDPNATTypeFunc(ctx, proxy, timeout)
	m.Info = info
	if err != nil && strings.TrimSpace(m.Info.Error) == "" {
		m.Info.Error = err.Error()
	}
	return nil
}

func resolveProbeTimeout(task *interfaces.Task) time.Duration {
	timeout := minProbeTimeout
	if task != nil {
		if raw := time.Duration(task.Config.Normalize().TaskTimeoutMillis) * time.Millisecond; raw > timeout {
			timeout = raw
		}
	}
	return timeout
}

func detectUDPNATType(ctx context.Context, proxy interfaces.Vendor, timeout time.Duration) (interfaces.UDPNATInfo, error) {
	if proxy == nil {
		return interfaces.UDPNATInfo{}, fmt.Errorf("proxy is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout < minProbeTimeout {
		timeout = minProbeTimeout
	}

	deadline := time.Now().Add(timeout)
	var errs []error
	for _, server := range stunServers {
		if err := ctx.Err(); err != nil {
			return interfaces.UDPNATInfo{Error: err.Error()}, err
		}
		dialCtx, cancel := context.WithTimeout(ctx, timeout)
		conn, err := proxy.DialUDP(dialCtx, stunURL(server))
		cancel()
		if err != nil {
			errs = append(errs, fmt.Errorf("%s:%d: %w", server.Host, server.Port, err))
			continue
		}

		internal := packetConnEndpoint(conn.LocalAddr())
		info := interfaces.UDPNATInfo{
			InternalIP:   internal.IP,
			InternalPort: internal.Port,
		}

		first, err := bindingTestFunc(ctx, conn, server, deadline, false, false)
		if err != nil {
			_ = conn.Close()
			if errors.Is(err, errSTUNTimeout) {
				continue
			}
			errs = append(errs, fmt.Errorf("%s:%d: %w", server.Host, server.Port, err))
			continue
		}

		info.PublicIP = first.External.IP
		info.PublicPort = first.External.Port

		kind, classifyErr := classifyTopology(ctx, conn, server, deadline, internal, *first)
		_ = conn.Close()
		info.Type = kind
		if classifyErr != nil {
			info.Error = classifyErr.Error()
			return info, classifyErr
		}
		if info.Type == "" {
			info.Type = interfaces.UDPNATTypeUnknown
		}
		return info, nil
	}

	if len(errs) > 0 {
		return interfaces.UDPNATInfo{Error: errors.Join(errs...).Error()}, fmt.Errorf("udp nat test failed: %w", errors.Join(errs...))
	}
	return interfaces.UDPNATInfo{Type: interfaces.UDPNATTypeBlocked}, nil
}

func classifyTopology(ctx context.Context, conn net.PacketConn, server stunServer, deadline time.Time, internal stunEndpoint, first stunResponse) (interfaces.UDPNATType, error) {
	if endpointsEqual(internal, first.External) {
		resp, err := bindingTestFunc(ctx, conn, server, deadline, true, true)
		if err != nil {
			if errors.Is(err, errSTUNTimeout) {
				return interfaces.UDPNATTypeUDPFirewall, nil
			}
			return "", err
		}
		if resp != nil {
			return interfaces.UDPNATTypeOpen, nil
		}
		return interfaces.UDPNATTypeUDPFirewall, nil
	}

	resp, err := bindingTestFunc(ctx, conn, server, deadline, true, true)
	if err == nil && resp != nil {
		return interfaces.UDPNATTypeFullCone, nil
	}
	if err != nil && !errors.Is(err, errSTUNTimeout) {
		return "", err
	}

	if first.Changed.IP == "" || first.Changed.Port == 0 {
		return "", fmt.Errorf("stun response missing changed address")
	}

	changedResp, err := bindingTestFunc(ctx, conn, stunServer{Host: first.Changed.IP, Port: first.Changed.Port}, deadline, false, false)
	if err != nil {
		return "", err
	}

	if !endpointsEqual(first.External, changedResp.External) {
		return interfaces.UDPNATTypeSymmetric, nil
	}

	portOnlyResp, err := bindingTestFunc(ctx, conn, stunServer{Host: first.Changed.IP, Port: first.Changed.Port}, deadline, false, true)
	if err != nil {
		if errors.Is(err, errSTUNTimeout) {
			return interfaces.UDPNATTypeRestrictedPort, nil
		}
		return "", err
	}
	if portOnlyResp != nil {
		return interfaces.UDPNATTypeRestrictedCone, nil
	}
	return interfaces.UDPNATTypeRestrictedPort, nil
}

var errSTUNTimeout = errors.New("stun timeout")

func bindingTest(ctx context.Context, conn net.PacketConn, server stunServer, deadline time.Time, changeIP bool, changePort bool) (*stunResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	addr, err := resolveUDPAddr(ctx, server)
	if err != nil {
		return nil, err
	}

	remaining := time.Until(deadline)
	if remaining <= 0 {
		return nil, errSTUNTimeout
	}

	request, token, err := buildBindingRequest(changeIP, changePort)
	if err != nil {
		return nil, err
	}

	attempts := maxRequestRetries
	if remaining < time.Duration(maxRequestRetries)*requestRetryWindow {
		attempts = 1
	}

	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		remaining = time.Until(deadline)
		if remaining <= 0 {
			return nil, errSTUNTimeout
		}
		if err := conn.SetDeadline(time.Now().Add(minDuration(requestRetryWindow, remaining))); err != nil {
			return nil, err
		}
		if _, err := conn.WriteTo(request, addr); err != nil {
			return nil, err
		}

		resp, err := readBindingResponse(ctx, conn, token)
		if err == nil {
			return resp, nil
		}
		if errors.Is(err, errSTUNTimeout) && attempt < attempts-1 && time.Until(deadline) > 0 {
			continue
		}
		return nil, err
	}

	return nil, errSTUNTimeout
}

func resolveUDPAddr(ctx context.Context, server stunServer) (*net.UDPAddr, error) {
	host := strings.Trim(server.Host, "[]")
	if ip := net.ParseIP(host); ip != nil {
		return &net.UDPAddr{IP: ip, Port: server.Port}, nil
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no address found for %s", server.Host)
	}
	return &net.UDPAddr{IP: addrs[0].IP, Port: server.Port}, nil
}

func buildBindingRequest(changeIP bool, changePort bool) ([]byte, [16]byte, error) {
	var token [16]byte
	binary.BigEndian.PutUint32(token[:4], stunMagicCookie)
	if _, err := rand.Read(token[4:]); err != nil {
		return nil, token, err
	}

	payload := make([]byte, 0, 8)
	if changeIP || changePort {
		payload = append(payload, 0x00, 0x03, 0x00, 0x04)
		var value uint32
		if changePort {
			value |= 0x02
		}
		if changeIP {
			value |= 0x04
		}
		payload = binary.BigEndian.AppendUint32(payload, value)
	}

	request := make([]byte, 20+len(payload))
	binary.BigEndian.PutUint16(request[0:2], stunBindingRequest)
	binary.BigEndian.PutUint16(request[2:4], uint16(len(payload)))
	copy(request[4:20], token[:])
	copy(request[20:], payload)
	return request, token, nil
}

func readBindingResponse(ctx context.Context, conn net.PacketConn, token [16]byte) (*stunResponse, error) {
	buf := make([]byte, 2048)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		n, _, err := conn.ReadFrom(buf)
		if err != nil {
			if timeoutErr(err) {
				return nil, errSTUNTimeout
			}
			return nil, err
		}
		if n < 20 {
			continue
		}
		resp, err := parseBindingResponse(buf[:n], token)
		if err != nil {
			continue
		}
		return resp, nil
	}
}

func parseBindingResponse(raw []byte, token [16]byte) (*stunResponse, error) {
	if len(raw) < 20 {
		return nil, fmt.Errorf("stun response too short")
	}
	if binary.BigEndian.Uint16(raw[0:2]) != stunBindingResponse {
		return nil, fmt.Errorf("unexpected stun message type")
	}
	length := int(binary.BigEndian.Uint16(raw[2:4]))
	if len(raw) < 20+length {
		return nil, fmt.Errorf("stun response truncated")
	}
	if !equalToken(raw[4:20], token[:]) {
		return nil, fmt.Errorf("stun transaction mismatch")
	}

	resp := &stunResponse{}
	attrs := raw[20 : 20+length]
	for offset := 0; offset+4 <= len(attrs); {
		attrType := binary.BigEndian.Uint16(attrs[offset : offset+2])
		attrLength := int(binary.BigEndian.Uint16(attrs[offset+2 : offset+4]))
		offset += 4
		if offset+attrLength > len(attrs) {
			return nil, fmt.Errorf("stun attribute truncated")
		}

		value := attrs[offset : offset+attrLength]
		offset += attrLength
		if padding := attrLength % 4; padding != 0 {
			offset += 4 - padding
		}

		switch attrType {
		case stunAttrMappedAddress:
			resp.External = parseMappedAddress(value)
		case stunAttrXORMappedAddress:
			resp.External = parseXORMappedAddress(value, token)
		case stunAttrSourceAddress, stunAttrResponseOrigin:
			resp.Source = parseMappedAddress(value)
		case stunAttrChangedAddress, stunAttrOtherAddress:
			resp.Changed = parseMappedAddress(value)
		}
	}

	if resp.External.IP == "" || resp.External.Port == 0 {
		return nil, fmt.Errorf("stun response missing mapped address")
	}
	if resp.Source.IP == "" || resp.Source.Port == 0 {
		resp.Source = resp.Changed
	}
	return resp, nil
}

func parseMappedAddress(value []byte) stunEndpoint {
	if len(value) < 8 {
		return stunEndpoint{}
	}
	family := value[1]
	port := int(binary.BigEndian.Uint16(value[2:4]))
	switch family {
	case 0x01:
		if len(value) < 8 {
			return stunEndpoint{}
		}
		return stunEndpoint{IP: net.IP(value[4:8]).String(), Port: port}
	case 0x02:
		if len(value) < 20 {
			return stunEndpoint{}
		}
		return stunEndpoint{IP: net.IP(value[4:20]).String(), Port: port}
	default:
		return stunEndpoint{}
	}
}

func parseXORMappedAddress(value []byte, token [16]byte) stunEndpoint {
	if len(value) < 8 {
		return stunEndpoint{}
	}

	family := value[1]
	port := int(binary.BigEndian.Uint16(value[2:4]) ^ uint16(stunMagicCookie>>16))
	switch family {
	case 0x01:
		if len(value) < 8 {
			return stunEndpoint{}
		}
		addr := make([]byte, 4)
		cookie := token[:4]
		for i := range addr {
			addr[i] = value[4+i] ^ cookie[i]
		}
		return stunEndpoint{IP: net.IP(addr).String(), Port: port}
	case 0x02:
		if len(value) < 20 {
			return stunEndpoint{}
		}
		addr := make([]byte, 16)
		for i := range addr {
			addr[i] = value[4+i] ^ token[i]
		}
		return stunEndpoint{IP: net.IP(addr).String(), Port: port}
	default:
		return stunEndpoint{}
	}
}

func packetConnEndpoint(addr net.Addr) stunEndpoint {
	if addr == nil {
		return stunEndpoint{}
	}
	switch typed := addr.(type) {
	case *net.UDPAddr:
		ip := ""
		if typed.IP != nil {
			ip = typed.IP.String()
		}
		return stunEndpoint{IP: ip, Port: typed.Port}
	default:
		host, port, err := net.SplitHostPort(addr.String())
		if err != nil {
			return stunEndpoint{}
		}
		portValue, err := net.LookupPort("udp", port)
		if err != nil {
			return stunEndpoint{}
		}
		return stunEndpoint{IP: strings.Trim(host, "[]"), Port: portValue}
	}
}

func endpointsEqual(left stunEndpoint, right stunEndpoint) bool {
	if left.Port == 0 || right.Port == 0 {
		return false
	}
	if left.Port != right.Port {
		return false
	}
	if left.IP == "" || right.IP == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(left.IP), strings.TrimSpace(right.IP))
}

func stunURL(server stunServer) string {
	return fmt.Sprintf("stun://%s:%d", server.Host, server.Port)
}

func timeoutErr(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func minDuration(left time.Duration, right time.Duration) time.Duration {
	switch {
	case left <= 0:
		return right
	case right <= 0:
		return left
	case left < right:
		return left
	default:
		return right
	}
}

func equalToken(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
