package profiles

import "testing"

func TestParseNodes(t *testing.T) {
	content := `
proxies:
  - name: hk-1
    type: vmess
    server: 1.1.1.1
    port: 443
  - type: trojan
    server: 2.2.2.2
    port: 443
`

	nodes, err := ParseNodes(content)
	if err != nil {
		t.Fatalf("ParseNodes() error = %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Name != "hk-1" {
		t.Fatalf("expected first node name hk-1, got %s", nodes[0].Name)
	}
	if nodes[1].Name != "node-2" {
		t.Fatalf("expected second node default name node-2, got %s", nodes[1].Name)
	}
}
