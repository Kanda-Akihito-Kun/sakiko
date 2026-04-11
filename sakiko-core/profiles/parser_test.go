package profiles

import "testing"

func TestParseNodes(t *testing.T) {
	content := `
proxies:
  - name: hk-1
    type: vmess
    server: 1.1.1.1
    port: 443
    udp: true
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
	if nodes[0].Protocol != "vmess" || nodes[0].Server != "1.1.1.1" || nodes[0].Port != "443" {
		t.Fatalf("expected first node metadata populated, got %+v", nodes[0])
	}
	if nodes[0].UDP == nil || !*nodes[0].UDP {
		t.Fatalf("expected first node udp=true, got %+v", nodes[0].UDP)
	}
	if nodes[1].Protocol != "trojan" || nodes[1].Server != "2.2.2.2" || nodes[1].Port != "443" {
		t.Fatalf("expected second node metadata populated, got %+v", nodes[1])
	}
	if nodes[1].UDP != nil {
		t.Fatalf("expected second node udp to be unknown, got %+v", nodes[1].UDP)
	}
}
