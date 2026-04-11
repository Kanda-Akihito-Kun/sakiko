package executor

import (
	"testing"

	"sakiko.local/sakiko-core/interfaces"
)

func TestPollItemRetriesFailedMatrices(t *testing.T) {
	original := executeNodeAttemptFunc
	defer func() {
		executeNodeAttemptFunc = original
	}()

	attempts := 0
	executeNodeAttemptFunc = func(task interfaces.Task, idx int, matrices []interfaces.MatrixEntry, macros []interfaces.MacroType) interfaces.EntryResult {
		attempts++
		if attempts == 1 {
			return interfaces.EntryResult{
				ProxyInfo: interfaces.ProxyInfo{Name: "node-1"},
				Error:     "PING: ping failed",
				Matrices: []interfaces.MatrixResult{
					{Type: interfaces.MatrixRTTPing, Payload: map[string]any{"value": uint64(0)}},
				},
			}
		}
		return interfaces.EntryResult{
			ProxyInfo: interfaces.ProxyInfo{Name: "node-1"},
			Matrices: []interfaces.MatrixResult{
				{Type: interfaces.MatrixRTTPing, Payload: map[string]any{"value": uint64(123)}},
			},
		}
	}

	item := &pollItem{
		id: "task-1",
		task: interfaces.Task{
			ID: "task-1",
			Config: interfaces.TaskConfig{
				TaskRetry: 2,
			},
			Matrices: []interfaces.MatrixEntry{
				{Type: interfaces.MatrixRTTPing},
			},
		},
		matrices: []interfaces.MatrixEntry{
			{Type: interfaces.MatrixRTTPing},
		},
		results: make([]interfaces.EntryResult, 1),
	}

	item.Yield(0, nil)

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
	if got := item.results[0].Error; got != "" {
		t.Fatalf("expected retry to recover error, got %q", got)
	}
	if len(item.results[0].Matrices) != 1 {
		t.Fatalf("expected 1 matrix, got %d", len(item.results[0].Matrices))
	}
}

func TestCountMatrixFailuresDetectsGeoErrors(t *testing.T) {
	failures := countMatrixFailures(
		[]interfaces.MatrixEntry{{Type: interfaces.MatrixInboundGeoIP}},
		interfaces.EntryResult{
			Matrices: []interfaces.MatrixResult{
				{Type: interfaces.MatrixInboundGeoIP, Payload: interfaces.GeoIPInfo{Error: "lookup failed"}},
			},
		},
	)
	if failures != 1 {
		t.Fatalf("expected 1 failure, got %d", failures)
	}
}

func TestNodeRetryAttemptsDisablesWholeNodeRetryForExpensiveMacros(t *testing.T) {
	task := interfaces.Task{
		Config: interfaces.TaskConfig{
			TaskRetry: 2,
		},
	}

	if got := nodeRetryAttempts(task, []interfaces.MacroType{interfaces.MacroPing, interfaces.MacroGeo}); got != 2 {
		t.Fatalf("expected connection-only task to keep retries, got %d", got)
	}
	if got := nodeRetryAttempts(task, []interfaces.MacroType{interfaces.MacroPing, interfaces.MacroSpeed}); got != 1 {
		t.Fatalf("expected speed task to disable whole-node retry, got %d", got)
	}
	if got := nodeRetryAttempts(task, []interfaces.MacroType{interfaces.MacroMedia}); got != 1 {
		t.Fatalf("expected media task to disable whole-node retry, got %d", got)
	}
}
