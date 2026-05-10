package app

import (
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func TestNormalizeTaskPresets(t *testing.T) {
	got := NormalizeTaskPresets(nil, "speed+ping,media")
	want := []string{"ping", "speed", "media"}
	if len(got) != len(want) {
		t.Fatalf("preset count = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("preset[%d] = %q, want %q: %#v", i, got[i], want[i], got)
		}
	}
}

func TestFormatPresetLabel(t *testing.T) {
	if got := FormatPresetLabel([]string{"full"}); got != "full" {
		t.Fatalf("full label = %q", got)
	}
	if got := FormatPresetLabel([]string{"speed", "ping"}); got != "ping+speed" {
		t.Fatalf("combined label = %q", got)
	}
	if got := FormatPresetLabel(nil); got != "ping" {
		t.Fatalf("empty label = %q", got)
	}
}

func TestBuildProfileTaskFromProfile(t *testing.T) {
	now := func() time.Time {
		return time.Date(2026, 5, 9, 10, 11, 12, 0, time.UTC)
	}
	profile := interfaces.Profile{
		ID:     "profile-1",
		Name:   "Example",
		Source: "https://example.test/sub",
		Nodes: []interfaces.Node{
			{Name: "a", Enabled: true},
			{Name: "b", Enabled: false},
		},
	}

	task, info, err := BuildProfileTaskFromProfile(profile, ProfileTaskRequest{Preset: "ping+speed"}, now)
	if err != nil {
		t.Fatalf("BuildProfileTaskFromProfile() error = %v", err)
	}
	if task.Name != "Example PING+SPEED 10:11:12" {
		t.Fatalf("task name = %q", task.Name)
	}
	if task.Context.Preset != "ping+speed" || info.PresetLabel != "ping+speed" {
		t.Fatalf("preset labels = task %q info %q", task.Context.Preset, info.PresetLabel)
	}
	if len(task.Nodes) != 1 || task.Nodes[0].Name != "a" {
		t.Fatalf("selected nodes = %#v", task.Nodes)
	}
	if len(task.Matrices) != 6 {
		t.Fatalf("matrix count = %d, want 6", len(task.Matrices))
	}
}
