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

func TestManagerRefreshBySearchAddsSearchQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("search"); got != "Tokyo" {
			t.Fatalf("expected search query Tokyo, got %q", got)
		}
		if _, err := fmt.Fprint(w, `[
			{
				"id": "1",
				"name": "Tokyo",
				"country": "Japan",
				"cc": "JP",
				"sponsor": "Example",
				"host": "tokyo.example.net",
				"url": "https://tokyo.example.net/speedtest/upload.php"
			}
		]`); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	manager := NewManager(Config{
		Client:       server.Client(),
		Endpoint:     server.URL + "/servers?engine=js&https_functional=true&limit=5",
		FetchTimeout: time.Second,
		CacheTTL:     time.Minute,
	})

	targets, err := manager.RefreshBySearch("Tokyo")
	if err != nil {
		t.Fatalf("refresh targets by search: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}
	if targets[1].City != "Tokyo" {
		t.Fatalf("expected Tokyo target, got %+v", targets[1])
	}
}

func TestManagerListBySearchUsesCachePerSearchKey(t *testing.T) {
	var requests []string
	now := time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Query().Get("search"))
		if _, err := fmt.Fprint(w, `[
			{
				"id": "42",
				"name": "Tokyo",
				"country": "Japan",
				"cc": "JP",
				"sponsor": "Cached",
				"host": "tokyo.example.net",
				"url": "https://tokyo.example.net/speedtest/upload.php"
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

	if _, err := manager.ListBySearch("Tokyo"); err != nil {
		t.Fatalf("first search list: %v", err)
	}
	if _, err := manager.ListBySearch("Tokyo"); err != nil {
		t.Fatalf("second search list: %v", err)
	}
	if _, err := manager.ListBySearch("Osaka"); err != nil {
		t.Fatalf("third search list: %v", err)
	}

	if len(requests) != 2 {
		t.Fatalf("expected 2 upstream requests, got %d", len(requests))
	}
	if requests[0] != "Tokyo" || requests[1] != "Osaka" {
		t.Fatalf("unexpected search requests: %#v", requests)
	}
}
