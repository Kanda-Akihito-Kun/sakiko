package executor

import (
	"context"
	"testing"

	"sakiko.local/sakiko-core/interfaces"
)

func TestPollItemRetriesFailedMatrices(t *testing.T) {
	original := executeNodeAttemptFunc
	defer func() {
		executeNodeAttemptFunc = original
	}()

	attempts := 0
	executeNodeAttemptFunc = func(ctx context.Context, task interfaces.Task, idx int, matrices []interfaces.MatrixEntry, macros []interfaces.MacroType, onUpdate func(interfaces.TaskActiveNode)) interfaces.EntryResult {
		_ = ctx
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

func TestPollItemEmitsActiveNodeUpdates(t *testing.T) {
	original := executeNodeAttemptFunc
	defer func() {
		executeNodeAttemptFunc = original
	}()

	var updates []interfaces.TaskActiveNode
	executeNodeAttemptFunc = func(ctx context.Context, task interfaces.Task, idx int, matrices []interfaces.MatrixEntry, macros []interfaces.MacroType, onUpdate func(interfaces.TaskActiveNode)) interfaces.EntryResult {
		_ = ctx
		if onUpdate != nil {
			onUpdate(interfaces.TaskActiveNode{
				Phase:    interfaces.TaskRuntimePhaseMacro,
				Macro:    interfaces.MacroSpeed,
				Matrices: []interfaces.MatrixType{interfaces.MatrixAverageSpeed, interfaces.MatrixMaxSpeed},
			})
		}
		return interfaces.EntryResult{
			ProxyInfo: interfaces.ProxyInfo{Name: "node-1", Address: "example.com:443"},
			Matrices: []interfaces.MatrixResult{
				{Type: interfaces.MatrixAverageSpeed, Payload: map[string]any{"value": uint64(123)}},
			},
		}
	}

	item := &pollItem{
		id: "task-1",
		task: interfaces.Task{
			ID: "task-1",
			Nodes: []interfaces.Node{
				{Name: "node-1", Server: "example.com"},
			},
		},
		matrices: []interfaces.MatrixEntry{
			{Type: interfaces.MatrixAverageSpeed},
		},
		macros:  []interfaces.MacroType{interfaces.MacroSpeed},
		results: make([]interfaces.EntryResult, 1),
		onUpdate: func(self *pollItem, activeNode interfaces.TaskActiveNode) {
			updates = append(updates, activeNode)
		},
	}

	item.Yield(0, nil)

	if len(updates) != 1 {
		t.Fatalf("expected 1 active-node update, got %d", len(updates))
	}
	if updates[0].NodeIndex != 0 {
		t.Fatalf("expected node index 0, got %d", updates[0].NodeIndex)
	}
	if updates[0].NodeName != "node-1" {
		t.Fatalf("expected node name node-1, got %q", updates[0].NodeName)
	}
	if updates[0].Attempt != 1 {
		t.Fatalf("expected attempt 1, got %d", updates[0].Attempt)
	}
	if updates[0].Macro != interfaces.MacroSpeed {
		t.Fatalf("expected speed macro, got %q", updates[0].Macro)
	}
}

func TestPollItemCancelCancelsTaskContext(t *testing.T) {
	item := (&pollItem{
		id:      "task-1",
		task:    interfaces.Task{ID: "task-1"},
		results: make([]interfaces.EntryResult, 1),
	}).Init().(*pollItem)

	item.Cancel()

	select {
	case <-item.ctx.Done():
	default:
		t.Fatalf("expected task context to be canceled")
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

func TestCountMatrixFailuresTreatsBlockedUDPNATAsMeasuredResult(t *testing.T) {
	failures := countMatrixFailures(
		[]interfaces.MatrixEntry{{Type: interfaces.MatrixUDPNATType}},
		interfaces.EntryResult{
			Matrices: []interfaces.MatrixResult{
				{Type: interfaces.MatrixUDPNATType, Payload: interfaces.UDPNATInfo{Type: interfaces.UDPNATTypeBlocked}},
			},
		},
	)
	if failures != 0 {
		t.Fatalf("expected blocked UDP NAT result to count as success, got %d failures", failures)
	}
}

func TestCountMatrixFailuresIgnoresUDPNATErrors(t *testing.T) {
	failures := countMatrixFailures(
		[]interfaces.MatrixEntry{{Type: interfaces.MatrixUDPNATType}},
		interfaces.EntryResult{
			Matrices: []interfaces.MatrixResult{
				{Type: interfaces.MatrixUDPNATType, Payload: interfaces.UDPNATInfo{Error: "stun response missing changed address"}},
			},
		},
	)
	if failures != 0 {
		t.Fatalf("expected UDP NAT probe error to be non-fatal, got %d failures", failures)
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
	if got := nodeRetryAttempts(task, []interfaces.MacroType{interfaces.MacroPing, interfaces.MacroUDP}); got != 2 {
		t.Fatalf("expected udp task to keep retries, got %d", got)
	}
	if got := nodeRetryAttempts(task, []interfaces.MacroType{interfaces.MacroPing, interfaces.MacroSpeed}); got != 1 {
		t.Fatalf("expected speed task to disable whole-node retry, got %d", got)
	}
	if got := nodeRetryAttempts(task, []interfaces.MacroType{interfaces.MacroMedia}); got != 1 {
		t.Fatalf("expected media task to disable whole-node retry, got %d", got)
	}
}
