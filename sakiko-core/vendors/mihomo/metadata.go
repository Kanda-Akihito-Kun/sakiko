package mihomo

import (
	"fmt"
	"net/url"
	"strconv"

	C "github.com/metacubex/mihomo/constant"
)

func urlToMetadata(rawURL string, network C.NetWork) (*C.Metadata, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	port := u.Port()
	if port == "" {
		switch u.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			port = "443"
		}
	}
	portValue, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("cannot parse port: %w", err)
	}

	addr := &C.Metadata{
		NetWork: network,
		Host:    u.Hostname(),
		DstPort: uint16(portValue),
	}
	return addr, nil
}
