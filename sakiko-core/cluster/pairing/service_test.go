package pairing

import (
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func TestCreateAndConsumeCode(t *testing.T) {
	service := New()

	code, err := service.CreateCode(interfaces.ClusterCreatePairingCodeRequest{
		KnightName: "Knight-A",
	})
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}
	if code.Code == "" {
		t.Fatalf("expected non-empty code")
	}

	consumed, err := service.ConsumeCode(code.Code)
	if err != nil {
		t.Fatalf("ConsumeCode() error = %v", err)
	}
	if consumed.Code != code.Code {
		t.Fatalf("expected consumed code %q, got %q", code.Code, consumed.Code)
	}

	if _, err := service.ConsumeCode(code.Code); err == nil {
		t.Fatalf("expected single-use pairing code rejection")
	}
}

func TestConsumeCodeRejectsExpiredCode(t *testing.T) {
	service := New()
	base := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return base }

	code, err := service.CreateCode(interfaces.ClusterCreatePairingCodeRequest{
		TTLSeconds: 1,
	})
	if err != nil {
		t.Fatalf("CreateCode() error = %v", err)
	}

	service.now = func() time.Time { return base.Add(2 * time.Second) }
	if _, err := service.ConsumeCode(code.Code); err == nil {
		t.Fatalf("expected expired code rejection")
	}
}
