package geo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/vendors/local"
)

func TestMacroRunCapturesInboundAndOutboundGeo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/":
			_, _ = fmt.Fprint(w, `{"success":true,"ip":"203.0.113.10","city":"Los Angeles","country":"United States","country_code":"US","connection":{"asn":64501,"org":"Exit Org","isp":"Exit ISP"}}`)
		case "/198.51.100.5":
			_, _ = fmt.Fprint(w, `{"success":true,"ip":"198.51.100.5","city":"Tokyo","country":"Japan","country_code":"JP","connection":{"asn":64500,"org":"Entry Org","isp":"Entry ISP"}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	previousOutbound := outboundLookupURL
	previousLookupPattern := ipLookupURLPattern
	outboundLookupURL = server.URL + "/"
	ipLookupURLPattern = server.URL + "/%s"
	defer func() {
		outboundLookupURL = previousOutbound
		ipLookupURLPattern = previousLookupPattern
	}()

	task := &interfaces.Task{
		Config: interfaces.TaskConfig{
			TaskTimeoutMillis: 1000,
		},
	}
	vendor := (&local.Vendor{}).Build(interfaces.Node{
		Name:    "node-1",
		Payload: "198.51.100.5",
	})

	macro := &Macro{}
	if err := macro.Run(vendor, task); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if macro.Inbound.Address != "198.51.100.5" {
		t.Fatalf("expected inbound address 198.51.100.5, got %q", macro.Inbound.Address)
	}
	if macro.Inbound.IP != "198.51.100.5" {
		t.Fatalf("expected inbound IP 198.51.100.5, got %q", macro.Inbound.IP)
	}
	if macro.Inbound.ASN != 64500 {
		t.Fatalf("expected inbound ASN 64500, got %d", macro.Inbound.ASN)
	}
	if macro.Inbound.ASOrganization != "Entry Org" {
		t.Fatalf("expected inbound AS organization Entry Org, got %q", macro.Inbound.ASOrganization)
	}
	if macro.Inbound.City != "Tokyo" {
		t.Fatalf("expected inbound city Tokyo, got %q", macro.Inbound.City)
	}

	if macro.Outbound.IP != "203.0.113.10" {
		t.Fatalf("expected outbound IP 203.0.113.10, got %q", macro.Outbound.IP)
	}
	if macro.Outbound.ASN != 64501 {
		t.Fatalf("expected outbound ASN 64501, got %d", macro.Outbound.ASN)
	}
	if macro.Outbound.ASOrganization != "Exit Org" {
		t.Fatalf("expected outbound AS organization Exit Org, got %q", macro.Outbound.ASOrganization)
	}
	if macro.Outbound.City != "Los Angeles" {
		t.Fatalf("expected outbound city Los Angeles, got %q", macro.Outbound.City)
	}
}

func TestExtractHost(t *testing.T) {
	cases := map[string]string{
		"1.2.3.4:443":         "1.2.3.4",
		"example.com:8443":    "example.com",
		"[2001:db8::1]:443":   "2001:db8::1",
		"2001:db8::1":         "2001:db8::1",
		"direct":              "",
		"  example.com:443  ": "example.com",
	}

	for input, expected := range cases {
		if actual := extractHost(input); actual != expected {
			t.Fatalf("extractHost(%q) = %q, want %q", input, actual, expected)
		}
	}
}
