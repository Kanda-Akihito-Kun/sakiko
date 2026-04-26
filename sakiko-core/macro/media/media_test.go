package media

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func TestRunProbeWithRetryRetriesFailedResult(t *testing.T) {
	t.Parallel()

	attempts := 0
	result := runProbeWithRetry(context.Background(), nil, mediaProbeAttemptTimeout, func(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
		_ = ctx
		_ = proxy
		attempts++
		if attempts == 1 {
			return finalizeResult(interfaces.MediaUnlockPlatformResult{
				Platform: interfaces.MediaUnlockPlatformNetflix,
				Name:     "Netflix",
				Status:   interfaces.MediaUnlockStatusFailed,
				Error:    "network connection",
			})
		}
		return finalizeResult(interfaces.MediaUnlockPlatformResult{
			Platform: interfaces.MediaUnlockPlatformNetflix,
			Name:     "Netflix",
			Status:   interfaces.MediaUnlockStatusYes,
			Region:   "US",
		})
	})

	if attempts != 2 {
		t.Fatalf("expected 2 probe attempts, got %d", attempts)
	}
	if result.Status != interfaces.MediaUnlockStatusYes {
		t.Fatalf("expected retry to return yes, got %q", result.Status)
	}
	if result.Display != "Yes (Region: US)" {
		t.Fatalf("expected display to be regenerated, got %q", result.Display)
	}
}

func TestDefaultMediaDisplaySupportsExpandedStatuses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  interfaces.MediaUnlockPlatformResult
		expect string
	}{
		{
			name: "web only",
			input: interfaces.MediaUnlockPlatformResult{
				Status: interfaces.MediaUnlockStatusWebOnly,
				Region: "US",
				Error:  "disallowed isp[1]",
			},
			expect: "Web Only (Disallowed ISP[1];Region: US)",
		},
		{
			name: "oversea only",
			input: interfaces.MediaUnlockPlatformResult{
				Status: interfaces.MediaUnlockStatusOverseaOnly,
				Region: "SG",
			},
			expect: "Oversea Only (Region: SG)",
		},
		{
			name: "unsupported",
			input: interfaces.MediaUnlockPlatformResult{
				Status: interfaces.MediaUnlockStatusUnsupported,
			},
			expect: "Unsupported",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := defaultMediaDisplay(tc.input); got != tc.expect {
				t.Fatalf("expected %q, got %q", tc.expect, got)
			}
		})
	}
}

func TestEvaluateChatGPTProbeHandlesDisallowedISPFromCFDetails(t *testing.T) {
	t.Parallel()

	body := []byte(`{"cf_details":"traffic policy (1)"}`)
	status, errText := evaluateChatGPTProbe("JP", true, body)
	if status != interfaces.MediaUnlockStatusWebOnly {
		t.Fatalf("expected web only status, got %q", status)
	}
	if errText != "disallowed isp[1]" {
		t.Fatalf("expected disallowed isp[1], got %q", errText)
	}
}

func TestEvaluateChatGPTProbeTreatsResolvedSupportedRegionAsAccessible(t *testing.T) {
	t.Parallel()

	status, errText := evaluateChatGPTProbe("JP", false, []byte(`{"ok":true}`))
	if status != interfaces.MediaUnlockStatusYes {
		t.Fatalf("expected yes status with supported region, got %q", status)
	}
	if errText != "" {
		t.Fatalf("expected empty error text, got %q", errText)
	}
}

func TestEvaluateChatGPTProbeKeepsUnknownRegionWithoutRedirectAsNo(t *testing.T) {
	t.Parallel()

	status, errText := evaluateChatGPTProbe("", false, []byte(`{"ok":true}`))
	if status != interfaces.MediaUnlockStatusNo {
		t.Fatalf("expected no status without region or redirect, got %q", status)
	}
	if errText != "" {
		t.Fatalf("expected empty error text, got %q", errText)
	}
}

func TestExtractChatGPTCFDetailsSupportsStructuredPayload(t *testing.T) {
	t.Parallel()

	body := []byte(`{"cf_details":{"reason":"(2)"}}`)
	if got := extractChatGPTCFDetails(body); got != `{"reason":"(2)"}` {
		t.Fatalf("expected serialized cf_details, got %q", got)
	}
}

func TestEvaluateClaudeSnapshotTreatsUnavailableAsBlocked(t *testing.T) {
	t.Parallel()

	status, errText := evaluateClaudeSnapshot(httpSnapshot{
		StatusCode: 200,
		FinalURL:   "https://claude.ai/unavailable",
	})
	if status != interfaces.MediaUnlockStatusNo {
		t.Fatalf("expected no status, got %q", status)
	}
	if errText != "" {
		t.Fatalf("expected empty error text, got %q", errText)
	}
}

func TestEvaluateClaudeSnapshotKeepsNonUnavailable4xxAccessible(t *testing.T) {
	t.Parallel()

	status, errText := evaluateClaudeSnapshot(httpSnapshot{
		StatusCode: 403,
		FinalURL:   "https://claude.ai/login",
		Body:       []byte("challenge page"),
	})
	if status != interfaces.MediaUnlockStatusYes {
		t.Fatalf("expected yes status, got %q", status)
	}
	if errText != "" {
		t.Fatalf("expected empty error text, got %q", errText)
	}
}

func TestEvaluateHuluSnapshotTreatsReachableAuthErrorsAsUnlocked(t *testing.T) {
	t.Parallel()

	status, errText, err := evaluateHuluSnapshot(httpSnapshot{
		StatusCode: 401,
		Body:       []byte(`{"error":{"name":"LOGIN_BLOCKED"}}`),
	})
	if err != nil {
		t.Fatalf("expected nil parse error, got %v", err)
	}
	if status != interfaces.MediaUnlockStatusYes {
		t.Fatalf("expected yes status, got %q", status)
	}
	if errText != "LOGIN_BLOCKED" {
		t.Fatalf("expected LOGIN_BLOCKED, got %q", errText)
	}
}

func TestEvaluateHuluSnapshotDetectsGeoBlockedTextResponse(t *testing.T) {
	t.Parallel()

	status, errText, err := evaluateHuluSnapshot(httpSnapshot{
		StatusCode: 403,
		Body:       []byte("Hulu is available in the US only and cannot be used with an anonymous proxy service."),
	})
	if err != nil {
		t.Fatalf("expected nil parse error, got %v", err)
	}
	if status != interfaces.MediaUnlockStatusNo {
		t.Fatalf("expected no status, got %q", status)
	}
	if errText != "" {
		t.Fatalf("expected empty error text, got %q", errText)
	}
}

func TestHuluLoginPageReachableDetectsAccessibleLoginPage(t *testing.T) {
	t.Parallel()

	if !huluLoginPageReachable(httpSnapshot{
		StatusCode: 200,
		FinalURL:   "https://auth.hulu.com/web/login",
		Body:       []byte("<html><title>Hulu</title><body>Log In</body></html>"),
	}) {
		t.Fatal("expected login page to be considered reachable")
	}
}

func TestResolveProbeTimeoutRespectsTaskTimeoutWithinBounds(t *testing.T) {
	t.Parallel()

	task := &interfaces.Task{
		Config: interfaces.TaskConfig{
			TaskTimeoutMillis: 9000,
		},
	}

	if got := resolveProbeTimeout(task); got != 6*time.Second {
		t.Fatalf("expected capped 6s timeout, got %v", got)
	}
}

func TestResolveProbeTimeoutClampsLowValues(t *testing.T) {
	t.Parallel()

	task := &interfaces.Task{
		Config: interfaces.TaskConfig{
			TaskTimeoutMillis: 800,
		},
	}

	if got := resolveProbeTimeout(task); got != mediaProbeAttemptTimeout {
		t.Fatalf("expected min timeout %v, got %v", mediaProbeAttemptTimeout, got)
	}
}

func TestMediaProbeConcurrencyIsBounded(t *testing.T) {
	t.Parallel()

	var running atomic.Int32
	var maxRunning atomic.Int32
	probes := make([]func(context.Context, interfaces.Vendor) interfaces.MediaUnlockPlatformResult, 0, mediaProbeConcurrency+3)
	for range mediaProbeConcurrency + 3 {
		probes = append(probes, func(ctx context.Context, proxy interfaces.Vendor) interfaces.MediaUnlockPlatformResult {
			_ = ctx
			_ = proxy
			current := running.Add(1)
			for {
				previous := maxRunning.Load()
				if current <= previous || maxRunning.CompareAndSwap(previous, current) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			running.Add(-1)
			return finalizeResult(interfaces.MediaUnlockPlatformResult{
				Platform: interfaces.MediaUnlockPlatformNetflix,
				Name:     "Netflix",
				Status:   interfaces.MediaUnlockStatusYes,
				Region:   "US",
			})
		})
	}

	results := make([]interfaces.MediaUnlockPlatformResult, len(probes))
	sem := make(chan struct{}, mediaProbeConcurrency)
	done := make(chan struct{})
	for index, probe := range probes {
		go func(i int, run func(context.Context, interfaces.Vendor) interfaces.MediaUnlockPlatformResult) {
			sem <- struct{}{}
			defer func() {
				<-sem
				if i == len(probes)-1 {
					select {
					case done <- struct{}{}:
					default:
					}
				}
			}()
			results[i] = runProbeWithRetry(context.Background(), nil, mediaProbeAttemptTimeout, run)
		}(index, probe)
	}

	timeout := time.NewTimer(2 * time.Second)
	defer timeout.Stop()
	completed := 0
	for completed < len(probes) {
		select {
		case <-done:
			completed++
			if completed == len(probes) {
				break
			}
		case <-timeout.C:
			t.Fatalf("timed out waiting for bounded probes")
		default:
			allDone := true
			for _, result := range results {
				if result.Status == "" {
					allDone = false
					break
				}
			}
			if allDone {
				completed = len(probes)
			} else {
				time.Sleep(5 * time.Millisecond)
			}
		}
	}

	if got := maxRunning.Load(); got > mediaProbeConcurrency {
		t.Fatalf("expected at most %d concurrent probes, got %d", mediaProbeConcurrency, got)
	}
}
