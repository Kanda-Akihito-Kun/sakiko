package netx

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func RequestUnsafe(ctx context.Context, v interfaces.Vendor, opt interfaces.RequestOptions) (*http.Response, error) {
	method := opt.Method
	if method == "" {
		method = http.MethodGet
	}
	var body io.Reader
	if len(opt.Body) > 0 {
		body = bytes.NewReader(opt.Body)
	}
	req, err := http.NewRequestWithContext(ctx, method, opt.URL, body)
	if err != nil {
		return nil, err
	}
	if opt.OnConnected != nil {
		trace := &httptrace.ClientTrace{
			GotConn: func(info httptrace.GotConnInfo) {
				if info.Conn != nil {
					opt.OnConnected(info.Conn.RemoteAddr().String())
				}
			},
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	}
	if opt.Host != "" {
		req.Host = opt.Host
	}
	for key, value := range opt.Headers {
		req.Header.Set(key, value)
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
			if v == nil {
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			}
			return v.DialTCP(ctx, opt.URL, opt.Network)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: opt.TLSServerName,
		},
	}

	client := &http.Client{Transport: transport}
	if opt.NoRedir {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client.Do(req)
}
