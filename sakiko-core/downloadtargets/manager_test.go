package downloadtargets

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func TestManagerRefreshNormalizesTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/servers" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		if _, err := fmt.Fprint(w, `[
			{
				"id": "14623",
				"name": "Shanghai",
				"country": "China",
				"cc": "CN",
				"sponsor": "ExampleNet",
				"host": "shanghai.example.net",
				"url": "https://shanghai.example.net/speedtest/upload.php"
			}
		]`); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	manager := NewManager(Config{
		Client:       server.Client(),
		Endpoint:     server.URL + "/servers",
		FetchTimeout: time.Second,
		CacheTTL:     time.Minute,
	})

	targets, err := manager.Refresh()
	if err != nil {
		t.Fatalf("refresh targets: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}

	if targets[0].Source != interfaces.DownloadTargetSourceCloudflare {
		t.Fatalf("expected first target to be cloudflare, got %q", targets[0].Source)
	}

	speedtestTarget := targets[1]
	if speedtestTarget.Source != interfaces.DownloadTargetSourceSpeedtest {
		t.Fatalf("expected speedtest target, got %q", speedtestTarget.Source)
	}
	if speedtestTarget.DownloadURL != "https://shanghai.example.net/speedtest/random4000x4000.jpg" {
		t.Fatalf("unexpected download url %q", speedtestTarget.DownloadURL)
	}
	if speedtestTarget.Name != "Shanghai" {
		t.Fatalf("expected normalized target name, got %q", speedtestTarget.Name)
	}
}

func TestManagerListUsesCache(t *testing.T) {
	var requestCount int
	now := time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if _, err := fmt.Fprint(w, `[
			{
				"id": "42",
				"name": "Beijing",
				"country": "China",
				"cc": "CN",
				"sponsor": "Cached",
				"host": "bj.example.net",
				"url": "https://bj.example.net/speedtest/upload.php"
			}
		]`); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	manager := NewManager(Config{
		Client:       server.Client(),
		Endpoint:     server.URL,
		FetchTimeout: time.Second,
		CacheTTL:     time.Minute,
		Now: func() time.Time {
			return now
		},
	})

	if _, err := manager.List(); err != nil {
		t.Fatalf("first list: %v", err)
	}
	if _, err := manager.List(); err != nil {
		t.Fatalf("second list: %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("expected cached response, request count = %d", requestCount)
	}
}
