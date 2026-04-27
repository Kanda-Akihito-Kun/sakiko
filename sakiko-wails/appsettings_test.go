package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAppSettingsDefaultsPrivacyFlagsWhenFileMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")

	settings, err := loadAppSettings(path)
	if err != nil {
		t.Fatalf("loadAppSettings() error = %v", err)
	}
	if !settings.HideProfileNameInExport {
		t.Fatalf("expected HideProfileNameInExport default to be true")
	}
	if !settings.HideCNInboundInExport {
		t.Fatalf("expected HideCNInboundInExport default to be true")
	}
}

func TestLoadAppSettingsBackfillsMissingPrivacyFlags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"language":"en","dns":{"bootstrapServers":[],"resolverServers":[]}}`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	settings, err := loadAppSettings(path)
	if err != nil {
		t.Fatalf("loadAppSettings() error = %v", err)
	}
	if !settings.HideProfileNameInExport {
		t.Fatalf("expected HideProfileNameInExport to backfill to true")
	}
	if !settings.HideCNInboundInExport {
		t.Fatalf("expected HideCNInboundInExport to backfill to true")
	}
}
