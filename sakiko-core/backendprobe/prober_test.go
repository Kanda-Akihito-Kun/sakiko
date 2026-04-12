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

func TestBuildIPCNLocation(t *testing.T) {
	info := ipCNCurrentInfo{
		Country:  "中国",
		Province: "福建",
		City:     "福州",
		District: "闽侯",
		ISP:      "电信",
	}

	got := buildIPCNLocation(info)
	if got != "中国 福建 福州 闽侯 电信" {
		t.Fatalf("unexpected ip.cn location: %q", got)
	}
}

func TestExtractCurrentInfoFromIPCNDataObject(t *testing.T) {
	html := `
<script>
const state = {
  data: {
    ip: "59.61.129.169",
    country: "中国",
    province: "福建",
    city: "福州",
    district: "闽侯",
    isp: "电信"
  }
}
</script>`

	got, err := extractCurrentInfoFromIPCN(html)
	if err != nil {
		t.Fatalf("extractCurrentInfoFromIPCN() error = %v", err)
	}
	if got.IP != "59.61.129.169" {
		t.Fatalf("unexpected ip: %q", got.IP)
	}
	if buildIPCNLocation(got) != "中国 福建 福州 闽侯 电信" {
		t.Fatalf("unexpected location: %q", buildIPCNLocation(got))
	}
}

func TestExtractCurrentInfoFromIPCNHTMLLocation(t *testing.T) {
	html := `
<table>
  <tr>
    <td>所在地理位置</td>
    <td>中国 福建 福州 闽侯 电信</td>
  </tr>
</table>`

	got, err := extractCurrentInfoFromIPCN(html)
	if err != nil {
		t.Fatalf("extractCurrentInfoFromIPCN() error = %v", err)
	}
	if buildIPCNLocation(got) != "中国 福建 福州 闽侯 电信" {
		t.Fatalf("unexpected location: %q", buildIPCNLocation(got))
	}
}
