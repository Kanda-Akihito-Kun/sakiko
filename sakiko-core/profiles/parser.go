package profiles

import (
	"fmt"
	"strings"

	"sakiko.local/sakiko-core/interfaces"

	"github.com/metacubex/mihomo/common/convert"
	"go.yaml.in/yaml/v3"
)

type clashProfile struct {
	Proxies []map[string]any `yaml:"proxies"`
}

func ParseNodes(content string) ([]interfaces.Node, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, fmt.Errorf("profile content is empty")
	}

	var profile clashProfile
	if err := yaml.Unmarshal([]byte(trimmed), &profile); err == nil && len(profile.Proxies) > 0 {
		return mapsToNodes(profile.Proxies)
	}

	proxies, err := convert.ConvertsV2Ray([]byte(trimmed))
	if err == nil && len(proxies) > 0 {
		return mapsToNodes(proxies)
	}

	return nil, fmt.Errorf("profile content is neither clash yaml nor supported share-link subscription")
}

func mapsToNodes(proxies []map[string]any) ([]interfaces.Node, error) {
	nodes := make([]interfaces.Node, 0, len(proxies))
	for i, item := range proxies {
		name, _ := item["name"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			name = fmt.Sprintf("node-%d", i+1)
		}

		raw, err := yaml.Marshal(item)
		if err != nil {
			return nil, err
		}
		protocol := strings.TrimSpace(stringField(item["type"]))
		server := strings.TrimSpace(stringField(item["server"]))
		port := strings.TrimSpace(stringField(item["port"]))
		nodes = append(nodes, interfaces.Node{
			Name:     name,
			Protocol: protocol,
			Server:   server,
			Port:     port,
			UDP:      optionalBoolField(item["udp"]),
			Payload:  string(raw),
			Enabled:  true,
		})
	}
	return nodes, nil
}

func ComposeContent(nodes []interfaces.Node) (string, error) {
	proxies := make([]map[string]any, 0, len(nodes))
	for _, node := range nodes {
		raw := strings.TrimSpace(node.Payload)
		if raw == "" {
			continue
		}

		var item map[string]any
		if err := yaml.Unmarshal([]byte(raw), &item); err != nil {
			return "", err
		}
		proxies = append(proxies, item)
	}

	if len(proxies) == 0 {
		return "", fmt.Errorf("no valid node payloads to compose")
	}

	raw, err := yaml.Marshal(clashProfile{Proxies: proxies})
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func stringField(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case int:
		return fmt.Sprintf("%d", typed)
	case int8:
		return fmt.Sprintf("%d", typed)
	case int16:
		return fmt.Sprintf("%d", typed)
	case int32:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case uint:
		return fmt.Sprintf("%d", typed)
	case uint8:
		return fmt.Sprintf("%d", typed)
	case uint16:
		return fmt.Sprintf("%d", typed)
	case uint32:
		return fmt.Sprintf("%d", typed)
	case uint64:
		return fmt.Sprintf("%d", typed)
	case float32:
		return fmt.Sprintf("%.0f", typed)
	case float64:
		return fmt.Sprintf("%.0f", typed)
	default:
		return ""
	}
}

func optionalBoolField(value any) *bool {
	switch typed := value.(type) {
	case bool:
		return boolPtr(typed)
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "true", "yes", "on":
			return boolPtr(true)
		case "false", "no", "off":
			return boolPtr(false)
		default:
			return nil
		}
	default:
		return nil
	}
}

func boolPtr(value bool) *bool {
	flag := value
	return &flag
}
