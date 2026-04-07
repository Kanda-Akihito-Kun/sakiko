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
		nodes = append(nodes, interfaces.Node{
			Name:    name,
			Payload: string(raw),
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
