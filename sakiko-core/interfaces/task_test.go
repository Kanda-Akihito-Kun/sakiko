package interfaces

import "testing"

func TestTaskConfigNormalizeUsesDefaultBackendIdentity(t *testing.T) {
	cfg := (TaskConfig{}).Normalize()

	if cfg.BackendIdentity != defaultBackendIdentity {
		t.Fatalf("expected default backend identity %q, got %q", defaultBackendIdentity, cfg.BackendIdentity)
	}
}

func TestTaskConfigNormalizeClampsBackendIdentityToThirtyRunes(t *testing.T) {
	cfg := (TaskConfig{
		BackendIdentity: "1234567890123456789012345678901234567890",
	}).Normalize()

	if len([]rune(cfg.BackendIdentity)) != maxBackendIdentityRunes {
		t.Fatalf("expected backend identity length %d, got %d", maxBackendIdentityRunes, len([]rune(cfg.BackendIdentity)))
	}
}
