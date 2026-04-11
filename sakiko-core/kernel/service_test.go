package kernel

import (
	"strings"
	"sync"
	"testing"
	"time"

	"sakiko.local/sakiko-core/interfaces"
)

func TestRuntimeStatusIdle(t *testing.T) {
	svc, err := New(Config{Mode: interfaces.ModeSerial})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer svc.Stop()

	status := svc.RuntimeStatus()
	if status.Running {
		t.Fatalf("expected runtime to be idle")
	}
	if status.RunningTask != 0 {
		t.Fatalf("expected 0 running tasks, got %d", status.RunningTask)
	}
	if status.TotalTask != 0 {
		t.Fatalf("expected 0 total tasks, got %d", status.TotalTask)
	}
}

func TestGetTaskReturnsResultsAfterExit(t *testing.T) {
	svc, err := New(Config{Mode: interfaces.ModeSerial})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer svc.Stop()

	done := make(chan interfaces.Event, 1)
	taskID, err := svc.Submit(interfaces.Task{
		Name:   "invalid-vendor",
		Vendor: interfaces.VendorInvalid,
		Nodes: []interfaces.Node{
			{Name: "node-1"},
		},
	}, func(event interfaces.Event) {
		if event.Type == interfaces.EventExit {
			done <- event
		}
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for task exit")
	}

	resp, ok := svc.GetTask(taskID)
	if !ok {
		t.Fatalf("GetTask(%q) not found", taskID)
	}
	if resp.Task.Status != "finished" {
		t.Fatalf("expected finished task, got %s", resp.Task.Status)
	}
	if resp.ExitCode != "success" {
		t.Fatalf("expected success exit code, got %s", resp.ExitCode)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if resp.Results[0].Error != "vendor not ready" {
		t.Fatalf("expected vendor not ready error, got %q", resp.Results[0].Error)
	}

	status := svc.RuntimeStatus()
	if status.Running {
		t.Fatalf("expected runtime to be idle after task exit")
	}
	if status.TotalTask != 1 {
		t.Fatalf("expected 1 total task, got %d", status.TotalTask)
	}
}

func TestTaskCapturesMacroErrors(t *testing.T) {
	svc, err := New(Config{Mode: interfaces.ModeSerial})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer svc.Stop()

	done := make(chan interfaces.Event, 1)
	taskID, err := svc.Submit(interfaces.Task{
		Name:   "macro-error",
		Vendor: interfaces.VendorLocal,
		Nodes: []interfaces.Node{
			{Name: "node-1", Payload: "direct"},
		},
		Matrices: []interfaces.MatrixEntry{
			{Type: interfaces.MatrixRTTPing},
		},
		Config: interfaces.TaskConfig{
			PingAddress: "://bad",
			TaskRetry:   1,
		},
	}, func(event interfaces.Event) {
		if event.Type == interfaces.EventExit {
			done <- event
		}
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for task exit")
	}

	resp, ok := svc.GetTask(taskID)
	if !ok {
		t.Fatalf("GetTask(%q) not found", taskID)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(resp.Results))
	}
	if !strings.Contains(resp.Results[0].Error, "PING: ping failed") {
		t.Fatalf("expected ping failure, got %q", resp.Results[0].Error)
	}
}

func TestTaskArchiveWriterReceivesFinishedSnapshot(t *testing.T) {
	t.Parallel()

	writer := &captureArchiveWriter{}
	svc, err := New(Config{
		Mode:          interfaces.ModeSerial,
		ArchiveWriter: writer,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer svc.Stop()

	done := make(chan interfaces.Event, 1)
	taskID, err := svc.Submit(interfaces.Task{
		Name:   "archive-writer",
		Vendor: interfaces.VendorInvalid,
		Context: interfaces.TaskContext{
			Preset:      "ping",
			ProfileID:   "profile-1",
			ProfileName: "Profile 1",
		},
		Nodes: []interfaces.Node{
			{Name: "node-1"},
		},
	}, func(event interfaces.Event) {
		if event.Type == interfaces.EventExit {
			done <- event
		}
	})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timed out waiting for task exit")
	}

	snapshot, ok := writer.Last()
	if !ok {
		t.Fatalf("expected archive snapshot to be captured")
	}
	if snapshot.Task.ID != taskID {
		t.Fatalf("expected task ID %q, got %q", taskID, snapshot.Task.ID)
	}
	if snapshot.Task.Context.ProfileID != "profile-1" {
		t.Fatalf("expected profile-1, got %q", snapshot.Task.Context.ProfileID)
	}
	if snapshot.State.Status != "finished" {
		t.Fatalf("expected finished state, got %q", snapshot.State.Status)
	}
	if len(snapshot.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(snapshot.Results))
	}
}

func TestListTasksReturnsNewestFirst(t *testing.T) {
	t.Parallel()

	svc, err := New(Config{Mode: interfaces.ModeSerial})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer svc.Stop()

	svc.tasks = map[string]*taskRecord{
		"task-old": {
			state: interfaces.TaskState{
				TaskID:     "task-old",
				Name:       "old",
				Status:     "finished",
				StartedAt:  "2026-04-11T10:00:00Z",
				FinishedAt: "2026-04-11T10:02:00Z",
			},
		},
		"task-new": {
			state: interfaces.TaskState{
				TaskID:    "task-new",
				Name:      "new",
				Status:    "running",
				StartedAt: "2026-04-11T11:00:00Z",
			},
		},
		"task-mid": {
			state: interfaces.TaskState{
				TaskID:     "task-mid",
				Name:       "mid",
				Status:     "finished",
				StartedAt:  "2026-04-11T10:30:00Z",
				FinishedAt: "2026-04-11T10:31:00Z",
			},
		},
	}

	tasks := svc.ListTasks()
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(tasks))
	}

	if tasks[0].TaskID != "task-new" {
		t.Fatalf("expected newest task first, got %q", tasks[0].TaskID)
	}
	if tasks[1].TaskID != "task-mid" {
		t.Fatalf("expected middle task second, got %q", tasks[1].TaskID)
	}
	if tasks[2].TaskID != "task-old" {
		t.Fatalf("expected oldest task last, got %q", tasks[2].TaskID)
	}
}

type captureArchiveWriter struct {
	mu       sync.Mutex
	snapshot interfaces.TaskArchiveSnapshot
	ok       bool
}

func (w *captureArchiveWriter) SaveTaskArchive(snapshot interfaces.TaskArchiveSnapshot) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.snapshot = snapshot
	w.ok = true
	return nil
}

func (w *captureArchiveWriter) Last() (interfaces.TaskArchiveSnapshot, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.snapshot, w.ok
}
