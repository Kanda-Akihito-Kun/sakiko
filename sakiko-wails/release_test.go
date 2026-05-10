package main

import (
	"testing"
	"time"
)

func TestNormalizeReleaseVersion(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"0.1.0":        "v0.1.0",
		"v0.1.0":       "v0.1.0",
		"V0.1.1":       "v0.1.1",
		"v0.2.0-rc.1":  "v0.2.0-rc.1",
		"invalid-tag":  "",
		"":             "",
		"   v0.3.0   ": "v0.3.0",
	}

	for input, expected := range cases {
		if actual := normalizeReleaseVersion(input); actual != expected {
			t.Fatalf("normalizeReleaseVersion(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestCompareReleaseVersions(t *testing.T) {
	t.Parallel()

	if got := compareReleaseVersions("v0.1.1", "v0.1.0"); got <= 0 {
		t.Fatalf("expected newer version comparison to be positive, got %d", got)
	}
	if got := compareReleaseVersions("v0.1.0", "0.1.0"); got != 0 {
		t.Fatalf("expected equal version comparison to be zero, got %d", got)
	}
	if got := compareReleaseVersions("v0.1.0-rc.1", "v0.1.0"); got >= 0 {
		t.Fatalf("expected prerelease comparison to be negative, got %d", got)
	}
}

func TestBuildReleaseCheckResult(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 19, 11, 30, 0, 0, time.UTC)
	result := buildReleaseCheckResult(githubLatestReleaseResponse{
		TagName: "v0.2.0",
		Name:    "Sakiko v0.2.0",
		HTMLURL: "",
		Body:    "Bug fixes and polish.",
	}, now)

	if result.CurrentVersion != displayReleaseVersion(appVersion) {
		t.Fatalf("unexpected current version: %q", result.CurrentVersion)
	}
	if result.LatestVersion != "v0.2.0" {
		t.Fatalf("unexpected latest version: %q", result.LatestVersion)
	}
	if !result.HasUpdate {
		t.Fatal("expected hasUpdate to be true")
	}
	if result.ReleaseURL != releasePageURL {
		t.Fatalf("unexpected release URL: %q", result.ReleaseURL)
	}
	if result.CheckedAt != now.Format(time.RFC3339) {
		t.Fatalf("unexpected checkedAt: %q", result.CheckedAt)
	}
	if result.ReleaseNotes != "Bug fixes and polish." {
		t.Fatalf("unexpected release notes: %q", result.ReleaseNotes)
	}
}
