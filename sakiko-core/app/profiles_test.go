package app

import (
	"testing"

	"sakiko.local/sakiko-core/interfaces"
)

func TestBuildProfileSummarySubscriptionInfo(t *testing.T) {
	summary := BuildProfileSummary(interfaces.Profile{
		ID:        "profile-1",
		Name:      "Example",
		Source:    "local",
		UpdatedAt: "2026-05-09T00:00:00Z",
		Nodes:     []interfaces.Node{{Name: "a"}, {Name: "b"}},
		Attributes: map[string]any{
			"subscriptionUserinfo": map[string]any{
				"upload":   float64(100),
				"download": float64(200),
				"total":    float64(1000),
				"expire":   float64(1778284800),
			},
		},
	})

	if summary.RemainingBytes != 700 {
		t.Fatalf("remaining = %d", summary.RemainingBytes)
	}
	if summary.ExpiresAt != "2026-05-09T00:00:00Z" {
		t.Fatalf("expires at = %q", summary.ExpiresAt)
	}
	if summary.NodeCount != 2 {
		t.Fatalf("node count = %d", summary.NodeCount)
	}
}
