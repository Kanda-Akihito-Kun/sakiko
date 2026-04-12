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

func TestExtractLocationFromHTML(t *testing.T) {
	html := `
<table>
  <tr>
    <td>所在地理位置</td>
    <td>中国 福建 福州 闽侯 电信</td>
  </tr>
</table>`

	got, ok := extractLocationFromHTML(html)
	if !ok {
		t.Fatalf("expected location match")
	}
	if got != "中国 福建 福州 闽侯 电信" {
		t.Fatalf("unexpected extracted location: %q", got)
	}
}

func TestExtractLocationFromMojibakeHTML(t *testing.T) {
	html := `
<table>
  <tr>
    <td>鎵€鍦ㄥ湴鐞嗕綅缃?/td>
    <td>涓浗 绂忓缓 绂忓窞 闽侯 鐢典俊</td>
  </tr>
</table>`

	got, ok := extractLocationFromHTML(html)
	if !ok {
		t.Fatalf("expected mojibake location match")
	}
	if got != "涓浗 绂忓缓 绂忓窞 闽侯 鐢典俊" {
		t.Fatalf("unexpected extracted mojibake location: %q", got)
	}
}
