package interfaces

import "testing"

func TestTaskConfigNormalizeDownloadDefaultsAndBounds(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		cfg := (TaskConfig{}).Normalize()

		if cfg.DownloadURL != defaultDownloadURL {
			t.Fatalf("expected default download url %q, got %q", defaultDownloadURL, cfg.DownloadURL)
		}
		if cfg.DownloadDuration != defaultDownloadDuration {
			t.Fatalf("expected default download duration %d, got %d", defaultDownloadDuration, cfg.DownloadDuration)
		}
	})

	t.Run("clamps min", func(t *testing.T) {
		cfg := (TaskConfig{DownloadDuration: 3}).Normalize()

		if cfg.DownloadDuration != minDownloadDuration {
			t.Fatalf("expected min-clamped download duration %d, got %d", minDownloadDuration, cfg.DownloadDuration)
		}
	})

	t.Run("clamps max", func(t *testing.T) {
		cfg := (TaskConfig{DownloadDuration: 30}).Normalize()

		if cfg.DownloadDuration != maxDownloadDuration {
			t.Fatalf("expected max-clamped download duration %d, got %d", maxDownloadDuration, cfg.DownloadDuration)
		}
	})
}
