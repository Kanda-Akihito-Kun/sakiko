package ping

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	neturl "net/url"
	"strings"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/logx"

	"go.uber.org/zap"
)

type Macro struct {
	RTT   uint16
	Delay uint16
}

var (
	pingViaTraceFunc  = pingViaTrace
	pingViaNetcatFunc = pingViaNetcat
)

func (m *Macro) Type() interfaces.MacroType {
	return interfaces.MacroPing
}

func (m *Macro) Run(proxy interfaces.Vendor, task *interfaces.Task) error {
	cfg := task.Config.Normalize()
	rtt, delay := ping(proxy, cfg.PingAddress, cfg.PingAverageOver, int(cfg.TaskRetry), cfg.TaskTimeoutMillis)
	m.RTT = rtt
	m.Delay = delay
	if rtt == 0 && delay == 0 {
		return fmt.Errorf("ping failed")
	}
	return nil
}

func ping(proxy interfaces.Vendor, rawURL string, avg uint16, retry int, timeoutMS uint) (uint16, uint16) {
	if proxy == nil {
		return 0, 0
	}
	if avg == 0 {
		avg = 1
	}
	if retry < 1 {
		retry = 1
	}

	var rtts []uint16
	var delays []uint16
	for sampleIndex := 0; sampleIndex < int(avg); sampleIndex++ {
		rtt, delay, ok := retryPingSample(proxy, rawURL, timeoutMS, retry, sampleIndex+1, int(avg))
		if !ok {
			continue
		}
		rtts = append(rtts, rtt)
		delays = append(delays, delay)
	}
	if len(delays) == 0 {
		return 0, 0
	}
	return avgPositiveUint16(rtts), avgPositiveUint16(delays)
}

func retryPingSample(
	proxy interfaces.Vendor,
	rawURL string,
	timeoutMS uint,
	attempts int,
	sampleIndex int,
	sampleTotal int,
) (uint16, uint16, bool) {
	if attempts < 1 {
		attempts = 1
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMS)*time.Millisecond)
		rtt, delay, err := measurePingSample(ctx, proxy, rawURL)
		cancel()
		if err == nil && (rtt > 0 || delay > 0) {
			return rtt, delay, true
		}

		if attempt >= attempts {
			break
		}

		pingLogger().Info("retrying ping sample",
			zap.Int("sample", sampleIndex),
			zap.Int("sample_total", sampleTotal),
			zap.Int("attempt", attempt+1),
			zap.Int("max_attempts", attempts),
			zap.String("url", rawURL),
			zap.String("error", pingAttemptError(err, rtt, delay)),
		)
		time.Sleep(150 * time.Millisecond)
	}

	return 0, 0, false
}

func measurePingSample(ctx context.Context, proxy interfaces.Vendor, rawURL string) (uint16, uint16, error) {
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return pingViaTraceFunc(ctx, proxy, rawURL)
	}
	return pingViaNetcatFunc(ctx, proxy, rawURL)
}

func pingAttemptError(err error, rtt uint16, delay uint16) string {
	if err != nil {
		return err.Error()
	}
	if rtt == 0 && delay == 0 {
		return "empty ping result"
	}
	return ""
}

func pingViaTrace(ctx context.Context, proxy interfaces.Vendor, rawURL string) (uint16, uint16, error) {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			return proxy.DialTCP(ctx, rawURL, interfaces.ROptionsTCP)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       3 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, 0, err
	}

	var tlsStart, tlsEnd, writeStart, writeEnd int64
	trace := &httptrace.ClientTrace{
		TLSHandshakeStart: func() { tlsStart = time.Now().UnixMilli() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			if err == nil {
				tlsEnd = time.Now().UnixMilli()
			}
		},
		WroteHeaders: func() { writeStart = time.Now().UnixMilli() },
		GotFirstResponseByte: func() {
			writeEnd = time.Now().UnixMilli()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))

	start := time.Now().UnixMilli()
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	total := time.Now().UnixMilli() - start
	if strings.HasPrefix(rawURL, "https://") && tlsEnd > tlsStart && writeEnd > tlsEnd {
		return uint16(writeEnd - tlsEnd), uint16(total), nil
	}
	if writeStart > start && writeEnd > start {
		return uint16(writeEnd - writeStart), uint16(writeEnd - start), nil
	}
	return 0, uint16(total), nil
}

func pingViaNetcat(ctx context.Context, proxy interfaces.Vendor, rawURL string) (uint16, uint16, error) {
	u, err := neturl.Parse(rawURL)
	if err != nil {
		return 0, 0, err
	}
	path := u.EscapedPath()
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}
	payload := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", path, u.Host)

	start := time.Now().UnixMilli()
	conn, err := proxy.DialTCP(ctx, rawURL, interfaces.ROptionsTCP)
	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	reader := bufio.NewReader(conn)

	if _, err := conn.Write([]byte(payload)); err != nil {
		return 0, 0, err
	}
	firstByteAt := int64(0)
	if _, err := reader.ReadByte(); err != nil {
		return 0, 0, err
	}
	firstByteAt = time.Now().UnixMilli()
	return uint16(firstByteAt - start), uint16(firstByteAt - start), nil
}

func avgPositiveUint16(values []uint16) uint16 {
	var total uint64
	count := 0
	for _, value := range values {
		if value == 0 {
			continue
		}
		total += uint64(value)
		count++
	}
	if count == 0 {
		return 0
	}
	return uint16(total / uint64(count))
}

func pingLogger() *zap.Logger {
	return logx.Named("core.macro.ping")
}
