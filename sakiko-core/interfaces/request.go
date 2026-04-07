package interfaces

type RequestOptionsNetwork string

const (
	ROptionsTCP  RequestOptionsNetwork = "tcp"
	ROptionsTCP6 RequestOptionsNetwork = "tcp6"
)

func (n RequestOptionsNetwork) String() string {
	switch n {
	case ROptionsTCP:
		return "tcp"
	case ROptionsTCP6:
		return "tcp6"
	default:
		return "tcp"
	}
}

type RequestOptions struct {
	Method        string
	URL           string
	Host          string
	TLSServerName string
	OnConnected   func(string)
	Headers       map[string]string
	Body          []byte
	NoRedir       bool
	Network       RequestOptionsNetwork
}
