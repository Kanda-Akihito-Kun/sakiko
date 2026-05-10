package app

import (
	"os"
	"path/filepath"
	"testing"

	"sakiko.local/sakiko-core/interfaces"
)

func TestLoadSettingsDefaultsWhenMissing(t *testing.T) {
	settings, err := LoadSettings(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}
	if settings.Language != "en" {
		t.Fatalf("language = %q", settings.Language)
	}
	if !settings.HideProfileNameInExport || !settings.HideCNInboundInExport {
		t.Fatalf("privacy defaults not enabled: %#v", settings)
	}
}

func TestLoadSettingsNormalizesLanguageAndDNS(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	raw := []byte(`{"language":"zh-CN","dns":{"bootstrapServers":[" 1.1.1.1 "],"resolverServers":[" https://dns.example/dns-query "]}}`)
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	settings, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings() error = %v", err)
	}
	if settings.Language != "zh" {
		t.Fatalf("language = %q", settings.Language)
	}
	if len(settings.DNS.BootstrapServers) != 1 || settings.DNS.BootstrapServers[0] != "1.1.1.1" {
		t.Fatalf("bootstrap servers = %#v", settings.DNS.BootstrapServers)
	}
	if len(settings.DNS.ResolverServers) != 1 || settings.DNS.ResolverServers[0] != "https://dns.example/dns-query" {
		t.Fatalf("resolver servers = %#v", settings.DNS.ResolverServers)
	}
}

func TestApplySettingsPatch(t *testing.T) {
	hideProfile := false
	dns := interfaces.DNSConfig{BootstrapServers: []string{"9.9.9.9"}, ResolverServers: []string{"8.8.8.8"}}
	settings := ApplySettingsPatch(DefaultSettings(), SettingsPatch{
		Language:                "en-US",
		DNS:                     &dns,
		HideProfileNameInExport: &hideProfile,
	})

	if settings.Language != "en" {
		t.Fatalf("language = %q", settings.Language)
	}
	if settings.HideProfileNameInExport {
		t.Fatalf("hide profile flag was not patched")
	}
	if !settings.HideCNInboundInExport {
		t.Fatalf("unpatched hide CN flag changed")
	}
	if settings.DNS.BootstrapServers[0] != "9.9.9.9" {
		t.Fatalf("dns = %#v", settings.DNS)
	}
}
