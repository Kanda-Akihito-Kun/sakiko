package backendprobe

import "testing"

func TestBuildFallbackLocation(t *testing.T) {
	info := ipWhoIsResponse{
		Country: "China",
		City:    "Fuzhou",
	}
	info.Connection.ISP = "China Telecom"

	got := buildFallbackLocation(info)
	if got != "China Fuzhou China Telecom" {
		t.Fatalf("unexpected fallback location: %q", got)
	}
}

func TestLocationPatternExtractsIPCNLocation(t *testing.T) {
	html := `
<table>
  <tr>
    <td>所在地理位置</td>
    <td>中国 福建 福州 电信</td>
  </tr>
</table>`

	matches := locationPattern.FindStringSubmatch(html)
	if len(matches) < 2 {
		t.Fatalf("expected location match")
	}
	if matches[1] != "中国 福建 福州 电信" {
		t.Fatalf("unexpected matched location: %q", matches[1])
	}
}
