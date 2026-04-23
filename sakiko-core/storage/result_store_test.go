package storage

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/internal/testkit"
)

func TestResultStoreSaveLoadAndList(t *testing.T) {
	t.Parallel()

	_, storePath := testkit.TempProfilesStore(t)
	store := NewResultStore(storePath)
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
				{Type: interfaces.MatrixTrafficUsed},
			},
			Config: interfaces.TaskConfig{
				PingAddress:       "https://cp.cloudflare.com/generate_204",
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
					{Type: interfaces.MatrixTrafficUsed, Payload: map[string]any{"value": uint64(12_345_678)}},
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
	if got := archive.Report.Sections[0].Rows[0]["trafficUsedBytes"]; got != float64(12_345_678) {
		t.Fatalf("expected trafficUsedBytes 12345678, got %#v", got)
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

	_, storePath := testkit.TempProfilesStore(t)
	store := NewResultStore(storePath)
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

	testkit.MustWriteString(t, store.Path("task-sidecar"), "{not-json")

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

	_, storePath := testkit.TempProfilesStore(t)
	store := NewResultStore(storePath)
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

func TestBuildSpeedSectionPreservesTaskNodeOrder(t *testing.T) {
	t.Parallel()

	section := buildSpeedSection(interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			Context: interfaces.TaskContext{Preset: "speed"},
		},
		Results: []interfaces.EntryResult{
			{
				ProxyInfo: interfaces.ProxyInfo{Name: "node-a", Type: interfaces.ProxyShadowsocks},
				Matrices: []interfaces.MatrixResult{
					{Type: interfaces.MatrixAverageSpeed, Payload: uint64(1_000)},
					{Type: interfaces.MatrixMaxSpeed, Payload: uint64(2_000)},
				},
			},
			{
				ProxyInfo: interfaces.ProxyInfo{Name: "node-b", Type: interfaces.ProxyShadowsocks},
				Matrices: []interfaces.MatrixResult{
					{Type: interfaces.MatrixAverageSpeed, Payload: uint64(9_000)},
					{Type: interfaces.MatrixMaxSpeed, Payload: uint64(10_000)},
				},
			},
		},
	})

	if len(section.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(section.Rows))
	}
	if got := section.Rows[0]["nodeName"]; got != "node-a" {
		t.Fatalf("expected first row node-a, got %#v", got)
	}
	if got := section.Rows[1]["nodeName"]; got != "node-b" {
		t.Fatalf("expected second row node-b, got %#v", got)
	}
	if got := section.Rows[0]["rank"]; got != 1 {
		t.Fatalf("expected first row rank 1, got %#v", got)
	}
	if got := section.Rows[1]["rank"]; got != 2 {
		t.Fatalf("expected second row rank 2, got %#v", got)
	}
}

func TestBuildSpeedSectionMarksZeroTrafficRowFailed(t *testing.T) {
	t.Parallel()

	section := buildSpeedSection(interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			Context: interfaces.TaskContext{Preset: "speed"},
		},
		Results: []interfaces.EntryResult{
			{
				ProxyInfo: interfaces.ProxyInfo{Name: "node-a", Type: interfaces.ProxyShadowsocks},
				Matrices: []interfaces.MatrixResult{
					{Type: interfaces.MatrixAverageSpeed, Payload: map[string]any{"value": uint64(0)}},
					{Type: interfaces.MatrixMaxSpeed, Payload: map[string]any{"value": uint64(0)}},
					{Type: interfaces.MatrixTrafficUsed, Payload: map[string]any{"value": uint64(0)}},
				},
			},
		},
	})

	if got := section.Rows[0]["error"]; got != "Failed" {
		t.Fatalf("expected zero-traffic row status Failed, got %#v", got)
	}
	if got := section.Summary["successCount"]; got != 0 {
		t.Fatalf("expected successCount 0 for zero-traffic row, got %#v", got)
	}
}

func TestBuildSpeedSectionIncludesUDPNATTypeAfterProtocol(t *testing.T) {
	t.Parallel()

	section := buildSpeedSection(interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			Context: interfaces.TaskContext{Preset: "full"},
			Matrices: []interfaces.MatrixEntry{
				{Type: interfaces.MatrixAverageSpeed},
				{Type: interfaces.MatrixUDPNATType},
			},
		},
		Results: []interfaces.EntryResult{
			{
				ProxyInfo: interfaces.ProxyInfo{Name: "node-a", Type: interfaces.ProxyShadowsocks},
				Matrices: []interfaces.MatrixResult{
					{Type: interfaces.MatrixAverageSpeed, Payload: map[string]any{"value": uint64(1024)}},
					{Type: interfaces.MatrixUDPNATType, Payload: interfaces.UDPNATInfo{Type: interfaces.UDPNATTypeFullCone}},
				},
			},
		},
	})

	if len(section.Columns) < 4 {
		t.Fatalf("expected speed section columns, got %d", len(section.Columns))
	}
	if section.Columns[2].Key != "proxyType" || section.Columns[3].Key != "udpNatType" {
		t.Fatalf("expected udpNatType after proxyType, got %q then %q", section.Columns[2].Key, section.Columns[3].Key)
	}
	if got := section.Rows[0]["udpNatType"]; got != string(interfaces.UDPNATTypeFullCone) {
		t.Fatalf("expected full-cone NAT cell, got %#v", got)
	}
}

func TestBuildResultReportInlinesUDPNATWithSpeedSection(t *testing.T) {
	t.Parallel()

	report := buildResultReport(interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			Matrices: []interfaces.MatrixEntry{
				{Type: interfaces.MatrixAverageSpeed},
				{Type: interfaces.MatrixUDPNATType},
			},
		},
		Results: []interfaces.EntryResult{
			{
				ProxyInfo: interfaces.ProxyInfo{Name: "node-a", Type: interfaces.ProxyShadowsocks},
				Matrices: []interfaces.MatrixResult{
					{Type: interfaces.MatrixAverageSpeed, Payload: map[string]any{"value": uint64(1024)}},
					{Type: interfaces.MatrixUDPNATType, Payload: interfaces.UDPNATInfo{Type: interfaces.UDPNATTypeBlocked}},
				},
			},
		},
	})

	if len(report.Sections) != 1 {
		t.Fatalf("expected only speed section, got %d sections", len(report.Sections))
	}
	if report.Sections[0].Kind != "speed_table" {
		t.Fatalf("expected UDP NAT to be inlined into speed table, got %q", report.Sections[0].Kind)
	}
	if got := report.Sections[0].Rows[0]["udpNatType"]; got != string(interfaces.UDPNATTypeBlocked) {
		t.Fatalf("expected blocked UDP NAT cell, got %#v", got)
	}
}

func TestBuildMediaUnlockSectionUsesNodePlatformMatrix(t *testing.T) {
	t.Parallel()

	section := buildMediaUnlockSection(interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			Context: interfaces.TaskContext{Preset: "media"},
		},
		Results: []interfaces.EntryResult{
			{
				ProxyInfo: interfaces.ProxyInfo{
					Name: "node-a",
					Type: interfaces.ProxyShadowsocks,
				},
				Matrices: []interfaces.MatrixResult{
					{
						Type: interfaces.MatrixMediaUnlock,
						Payload: interfaces.MediaUnlockResult{
							Items: []interfaces.MediaUnlockPlatformResult{
								{
									Platform: interfaces.MediaUnlockPlatformChatGPT,
									Name:     "ChatGPT",
									Status:   interfaces.MediaUnlockStatusYes,
									Display:  "Unlocked (US)",
								},
								{
									Platform: interfaces.MediaUnlockPlatformNetflix,
									Name:     "Netflix",
									Status:   interfaces.MediaUnlockStatusOriginalsOnly,
									Display:  "Originals Only (US)",
								},
							},
						},
					},
				},
			},
		},
	})

	if section.Kind != "media_unlock_table" {
		t.Fatalf("expected media_unlock_table section, got %q", section.Kind)
	}
	if len(section.Columns) < 4 {
		t.Fatalf("expected matrix columns, got %d", len(section.Columns))
	}
	if section.Columns[0].Key != "nodeName" || section.Columns[1].Key != "proxyType" {
		t.Fatalf("expected fixed node columns first, got %q and %q", section.Columns[0].Key, section.Columns[1].Key)
	}
	if len(section.Rows) != 1 {
		t.Fatalf("expected 1 matrix row, got %d", len(section.Rows))
	}
	if got := section.Rows[0]["chatgpt"]; got != "Unlocked (US)" {
		t.Fatalf("expected chatgpt cell to use display text, got %#v", got)
	}
	if got := section.Rows[0]["netflix"]; got != "Originals Only (US)" {
		t.Fatalf("expected netflix cell to use display text, got %#v", got)
	}
}

func TestBuildMediaUnlockSectionFiltersRemovedPlatforms(t *testing.T) {
	t.Parallel()

	section := buildMediaUnlockSection(interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			Context: interfaces.TaskContext{Preset: "media"},
		},
		Results: []interfaces.EntryResult{
			{
				ProxyInfo: interfaces.ProxyInfo{
					Name: "node-a",
					Type: interfaces.ProxyShadowsocks,
				},
				Matrices: []interfaces.MatrixResult{
					{
						Type: interfaces.MatrixMediaUnlock,
						Payload: interfaces.MediaUnlockResult{
							Items: []interfaces.MediaUnlockPlatformResult{
								{
									Platform: "dazn",
									Name:     "DAZN",
									Status:   interfaces.MediaUnlockStatusYes,
									Display:  "Unlocked (JP)",
								},
								{
									Platform: "instagram_music",
									Name:     "Instagram Music",
									Status:   interfaces.MediaUnlockStatusNo,
									Display:  "No",
								},
								{
									Platform: interfaces.MediaUnlockPlatformHuluJP,
									Name:     "Hulu Japan",
									Status:   interfaces.MediaUnlockStatusYes,
									Display:  "Unlocked (JP)",
								},
								{
									Platform: interfaces.MediaUnlockPlatformSpotify,
									Name:     "Spotify",
									Status:   interfaces.MediaUnlockStatusYes,
									Display:  "Region (US)",
								},
								{
									Platform: interfaces.MediaUnlockPlatformSteam,
									Name:     "Steam",
									Status:   interfaces.MediaUnlockStatusYes,
									Display:  "Currency (USD)",
								},
								{
									Platform: interfaces.MediaUnlockPlatformNetflix,
									Name:     "Netflix",
									Status:   interfaces.MediaUnlockStatusYes,
									Display:  "Unlocked (US)",
								},
							},
						},
					},
				},
			},
		},
	})

	for _, column := range section.Columns {
		if column.Key == "dazn" || column.Key == "instagram_music" || column.Key == "hulu_jp" || column.Key == "spotify" || column.Key == "steam" {
			t.Fatalf("expected removed media columns to be hidden, got %q", column.Key)
		}
	}
	if _, exists := section.Rows[0]["dazn"]; exists {
		t.Fatalf("expected DAZN cell to be omitted from report row")
	}
	if _, exists := section.Rows[0]["instagram_music"]; exists {
		t.Fatalf("expected Instagram Music cell to be omitted from report row")
	}
	if _, exists := section.Rows[0]["hulu_jp"]; exists {
		t.Fatalf("expected Hulu JP cell to be omitted from report row")
	}
	if _, exists := section.Rows[0]["spotify"]; exists {
		t.Fatalf("expected Spotify cell to be omitted from report row")
	}
	if _, exists := section.Rows[0]["steam"]; exists {
		t.Fatalf("expected Steam cell to be omitted from report row")
	}
	if got := section.Rows[0]["netflix"]; got != "Unlocked (US)" {
		t.Fatalf("expected visible platform to remain, got %#v", got)
	}
	if got := section.Summary["platformCount"]; got != 1 {
		t.Fatalf("expected platformCount 1 after filtering, got %#v", got)
	}
	if got := section.Summary["successCount"]; got != 1 {
		t.Fatalf("expected successCount 1 after filtering, got %#v", got)
	}
}

func TestBuildTopologySectionUsesCompactASNAndInfoColumns(t *testing.T) {
	t.Parallel()

	section := buildTopologySection(interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			Context: interfaces.TaskContext{Preset: "geo"},
		},
		Results: []interfaces.EntryResult{
			{
				ProxyInfo: interfaces.ProxyInfo{Name: "node-a", Type: interfaces.ProxyShadowsocks},
				Matrices: []interfaces.MatrixResult{
					{Type: interfaces.MatrixInboundGeoIP, Payload: interfaces.GeoIPInfo{IP: "59.61.129.169", ASN: 9808, ASOrganization: "China Mobile", City: "Wuhan"}},
					{Type: interfaces.MatrixOutboundGeoIP, Payload: interfaces.GeoIPInfo{IP: "203.0.113.8", ASN: 2516, ASOrganization: "KDDI", City: "Tokyo"}},
				},
			},
		},
	})

	columnKeys := make([]string, 0, len(section.Columns))
	for _, column := range section.Columns {
		columnKeys = append(columnKeys, column.Key)
	}
	expected := []string{"nodeName", "proxyType", "inboundASN", "inboundIP", "inboundInfo", "outboundASN", "outboundIP", "outboundInfo", "error"}
	if !reflect.DeepEqual(columnKeys, expected) {
		t.Fatalf("expected compact topology columns %#v, got %#v", expected, columnKeys)
	}
	if got := section.Rows[0]["inboundASN"]; got != 9808 {
		t.Fatalf("expected inbound ASN 9808, got %#v", got)
	}
	if got := section.Rows[0]["inboundIP"]; got != "59.61.129.169" {
		t.Fatalf("expected inbound IP, got %#v", got)
	}
	if got := section.Rows[0]["inboundInfo"]; got != "Wuhan | China Mobile" {
		t.Fatalf("expected inbound info, got %#v", got)
	}
	if got := section.Rows[0]["outboundASN"]; got != 2516 {
		t.Fatalf("expected outbound ASN 2516, got %#v", got)
	}
	if got := section.Rows[0]["outboundIP"]; got != "203.0.113.8" {
		t.Fatalf("expected outbound IP, got %#v", got)
	}
	if got := section.Rows[0]["outboundInfo"]; got != "Tokyo | KDDI" {
		t.Fatalf("expected outbound info, got %#v", got)
	}
}

func TestBuildUDPNATSectionUsesEndpointSummary(t *testing.T) {
	t.Parallel()

	section := buildUDPNATSection(interfaces.TaskArchiveSnapshot{
		Task: interfaces.Task{
			Context: interfaces.TaskContext{Preset: "udp"},
		},
		Results: []interfaces.EntryResult{
			{
				ProxyInfo: interfaces.ProxyInfo{
					Name: "node-a",
					Type: interfaces.ProxyShadowsocks,
				},
				Matrices: []interfaces.MatrixResult{
					{
						Type: interfaces.MatrixUDPNATType,
						Payload: interfaces.UDPNATInfo{
							Type:         interfaces.UDPNATTypeFullCone,
							InternalIP:   "10.0.0.2",
							InternalPort: 40000,
							PublicIP:     "198.51.100.8",
							PublicPort:   50000,
						},
					},
				},
			},
		},
	})

	if section.Kind != "udp_nat_table" {
		t.Fatalf("expected udp_nat_table section, got %q", section.Kind)
	}
	if len(section.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(section.Rows))
	}
	if got := section.Rows[0]["natType"]; got != interfaces.UDPNATTypeFullCone {
		t.Fatalf("expected full-cone NAT, got %#v", got)
	}
	if got := section.Rows[0]["internalEndpoint"]; got != "10.0.0.2:40000" {
		t.Fatalf("expected internal endpoint, got %#v", got)
	}
	if got := section.Rows[0]["publicEndpoint"]; got != "198.51.100.8:50000" {
		t.Fatalf("expected public endpoint, got %#v", got)
	}
	if got := section.Summary["successCount"]; got != 1 {
		t.Fatalf("expected successCount 1, got %#v", got)
	}
}
