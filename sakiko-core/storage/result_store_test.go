package storage

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"sakiko.local/sakiko-core/interfaces"
)

func TestResultStoreSaveLoadAndList(t *testing.T) {
	t.Parallel()

	store := NewResultStore(t.TempDir() + "/profiles.yaml")
	snapshot := interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			ID:     "task-1",
			Name:   "Example SPEED",
			Vendor: interfaces.VendorMihomo,
			Context: interfaces.TaskContext{
				Preset:      "speed",
				ProfileID:   "profile-1",
				ProfileName: "Example Profile",
			},
			Nodes: []interfaces.Node{
				{Name: "node-a", Payload: "type: shadowsocks"},
			},
			Matrices: []interfaces.MatrixEntry{
				{Type: interfaces.MatrixRTTPing},
				{Type: interfaces.MatrixHTTPPing},
				{Type: interfaces.MatrixAverageSpeed},
				{Type: interfaces.MatrixMaxSpeed},
				{Type: interfaces.MatrixPerSecSpeed},
			},
			Config: interfaces.TaskConfig{
				PingAddress:       "https://www.gstatic.com/generate_204",
				TaskTimeoutMillis: 6000,
				DownloadURL:       "https://speed.cloudflare.com/__down?bytes=10000000",
				DownloadDuration:  10,
				DownloadThreading: 8,
			},
		},
		State: interfaces.TaskState{
			TaskID:     "task-1",
			Name:       "Example SPEED",
			Status:     "finished",
			Progress:   1,
			Total:      1,
			StartedAt:  "2026-04-07T10:00:00Z",
			FinishedAt: "2026-04-07T10:01:00Z",
		},
		Results: []interfaces.EntryResult{
			{
				ProxyInfo: interfaces.ProxyInfo{
					Name:    "node-a",
					Address: "1.1.1.1:443",
					Type:    interfaces.ProxyShadowsocks,
				},
				Matrices: []interfaces.MatrixResult{
					{Type: interfaces.MatrixRTTPing, Payload: map[string]any{"value": uint64(123)}},
					{Type: interfaces.MatrixHTTPPing, Payload: map[string]any{"value": uint64(456)}},
					{Type: interfaces.MatrixAverageSpeed, Payload: map[string]any{"value": uint64(789)}},
					{Type: interfaces.MatrixMaxSpeed, Payload: map[string]any{"value": uint64(999)}},
					{Type: interfaces.MatrixPerSecSpeed, Payload: map[string]any{"values": []uint64{111, 222}}},
				},
			},
		},
		ExitCode: "success",
	}

	if err := store.SaveTaskArchive(snapshot); err != nil {
		t.Fatalf("SaveTaskArchive() error = %v", err)
	}

	archive, err := store.Load("task-1")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if archive.Task.ID != "task-1" {
		t.Fatalf("expected task ID task-1, got %q", archive.Task.ID)
	}
	if archive.Task.Context.Preset != "speed" {
		t.Fatalf("expected preset speed, got %q", archive.Task.Context.Preset)
	}
	if len(archive.Task.Nodes) != 1 {
		t.Fatalf("expected 1 archived node, got %d", len(archive.Task.Nodes))
	}
	if archive.Task.Nodes[0].Name != "node-a" {
		t.Fatalf("expected archived node name node-a, got %q", archive.Task.Nodes[0].Name)
	}

	raw, err := os.ReadFile(store.Path("task-1"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var stored map[string]any
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	taskMap, ok := stored["task"].(map[string]any)
	if !ok {
		t.Fatalf("expected task object in archive json")
	}
	nodes, ok := taskMap["nodes"].([]any)
	if !ok || len(nodes) != 1 {
		t.Fatalf("expected 1 node in archive json")
	}
	node, ok := nodes[0].(map[string]any)
	if !ok {
		t.Fatalf("expected archived node object")
	}
	if _, exists := node["payload"]; exists {
		t.Fatalf("expected archived node json to omit sensitive payload field")
	}

	if len(archive.Report.Sections) != 1 {
		t.Fatalf("expected 1 report section, got %d", len(archive.Report.Sections))
	}
	if archive.Report.Sections[0].Kind != "speed_table" {
		t.Fatalf("expected speed_table section, got %q", archive.Report.Sections[0].Kind)
	}

	items, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 archive list item, got %d", len(items))
	}
	if items[0].TaskID != "task-1" {
		t.Fatalf("expected task ID task-1, got %q", items[0].TaskID)
	}
}

func TestPreferredGeoOrganizationPrefersISPOverPrivateCustomer(t *testing.T) {
	t.Parallel()

	got := preferredGeoOrganization(interfaces.GeoIPInfo{
		ASOrganization: "Private Customer",
		ISP:            "Oracle Public Cloud",
	})
	if got != "Oracle Public Cloud" {
		t.Fatalf("expected Oracle Public Cloud, got %q", got)
	}
}

func TestResultStoreListUsesSummarySidecar(t *testing.T) {
	t.Parallel()

	store := NewResultStore(t.TempDir() + "/profiles.yaml")
	snapshot := interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			ID:   "task-sidecar",
			Name: "Sidecar SPEED",
			Context: interfaces.TaskContext{
				Preset:      "speed",
				ProfileName: "Sidecar Profile",
			},
			Nodes: []interfaces.Node{{Name: "node-a"}},
		},
		State: interfaces.TaskState{
			TaskID:     "task-sidecar",
			Name:       "Sidecar SPEED",
			Status:     "finished",
			StartedAt:  "2026-04-07T11:00:00Z",
			FinishedAt: "2026-04-07T11:01:00Z",
		},
		ExitCode: "success",
	}

	if err := store.SaveTaskArchive(snapshot); err != nil {
		t.Fatalf("SaveTaskArchive() error = %v", err)
	}

	if _, err := os.Stat(store.summaryPath("task-sidecar")); err != nil {
		t.Fatalf("expected summary sidecar to exist: %v", err)
	}

	if err := os.WriteFile(store.Path("task-sidecar"), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	items, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 archive list item, got %d", len(items))
	}
	if items[0].TaskID != "task-sidecar" {
		t.Fatalf("expected task-sidecar, got %q", items[0].TaskID)
	}
}

func TestResultStoreDeleteRemovesArchiveAndSummary(t *testing.T) {
	t.Parallel()

	store := NewResultStore(t.TempDir() + "/profiles.yaml")
	snapshot := interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			ID:   "task-delete",
			Name: "Delete SPEED",
			Context: interfaces.TaskContext{
				Preset:      "speed",
				ProfileName: "Delete Profile",
			},
			Nodes: []interfaces.Node{{Name: "node-a"}},
		},
		State: interfaces.TaskState{
			TaskID:     "task-delete",
			Name:       "Delete SPEED",
			Status:     "finished",
			StartedAt:  "2026-04-07T12:00:00Z",
			FinishedAt: "2026-04-07T12:01:00Z",
		},
		ExitCode: "success",
	}

	if err := store.SaveTaskArchive(snapshot); err != nil {
		t.Fatalf("SaveTaskArchive() error = %v", err)
	}

	if err := store.Delete("task-delete"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	if _, err := os.Stat(store.Path("task-delete")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected archive file to be deleted, got %v", err)
	}
	if _, err := os.Stat(store.summaryPath("task-delete")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected summary file to be deleted, got %v", err)
	}

	items, err := store.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no archived results after delete, got %d", len(items))
	}
}
