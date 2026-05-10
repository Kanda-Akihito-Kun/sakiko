package local

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func TestDialUDPReturnsPacketConnUsableWithWriteToAndReadFrom(t *testing.T) {
	server, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket() error = %v", err)
	}
	defer server.Close()

	serverDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 64)
		if err := server.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
			serverDone <- err
			return
		}

		n, addr, err := server.ReadFrom(buf)
		if err != nil {
			serverDone <- err
			return
		}
		if string(buf[:n]) != "ping" {
			serverDone <- &net.AddrError{Err: "unexpected payload", Addr: string(buf[:n])}
			return
		}
		_, err = server.WriteTo([]byte("pong"), addr)
		serverDone <- err
	}()

	vendor := (&Vendor{}).Build(interfaces.Node{Name: "direct"})
	conn, err := vendor.DialUDP(context.Background(), "stun://127.0.0.1:"+portOf(t, server.LocalAddr()))
	if err != nil {
		t.Fatalf("DialUDP() error = %v", err)
	}
	defer conn.Close()

	target, err := net.ResolveUDPAddr("udp4", server.LocalAddr().String())
	if err != nil {
		t.Fatalf("ResolveUDPAddr() error = %v", err)
	}

	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("SetDeadline() error = %v", err)
	}
	if _, err := conn.WriteTo([]byte("ping"), target); err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	buf := make([]byte, 64)
	n, _, err := conn.ReadFrom(buf)
	if err != nil {
		t.Fatalf("ReadFrom() error = %v", err)
	}
	if string(buf[:n]) != "pong" {
		t.Fatalf("expected pong, got %q", string(buf[:n]))
	}

	if err := <-serverDone; err != nil {
		t.Fatalf("server goroutine error = %v", err)
	}
}

func portOf(t *testing.T, addr net.Addr) string {
	t.Helper()

	udpAddr, ok := addr.(*net.UDPAddr)
	if !ok {
		t.Fatalf("expected UDPAddr, got %T", addr)
	}
	return strconv.Itoa(udpAddr.Port)
}
