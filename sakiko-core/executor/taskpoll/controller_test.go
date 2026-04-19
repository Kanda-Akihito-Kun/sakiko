package taskpoll

import (
	"sync"
	"testing"
	"time"
)

func TestControllerProcessesQueuedItemsSequentially(t *testing.T) {
	t.Parallel()

	controller := New("speed", 2, 0, time.Millisecond)
	stop := make(chan struct{})
	defer close(stop)

	go controller.Start(stop)

	first := newBlockingItem("first")
	second := newBlockingItem("second")

	controller.Push(first)
	controller.Push(second)

	select {
	case <-first.started:
	case <-time.After(time.Second):
		t.Fatalf("first item did not start")
	}

	select {
	case <-second.started:
		t.Fatalf("second item started before first item completed")
	case <-time.After(50 * time.Millisecond):
	}

	close(first.release)

	select {
	case <-second.started:
	case <-time.After(time.Second):
		t.Fatalf("second item did not start after first item completed")
	}

	close(second.release)
}

func TestControllerCancelQueuedItemExitsImmediately(t *testing.T) {
	t.Parallel()

	controller := New("speed", 1, 0, time.Millisecond)
	stop := make(chan struct{})
	defer close(stop)

	go controller.Start(stop)

	first := newBlockingItem("first")
	second := newBlockingItem("second")

	controller.Push(first)
	controller.Push(second)

	select {
	case <-first.started:
	case <-time.After(time.Second):
		t.Fatalf("first item did not start")
	}

	if ok := controller.Cancel("second"); !ok {
		t.Fatalf("expected queued second item to be cancelable")
	}

	select {
	case code := <-second.exitCode:
		if code != ExitCodeStopped {
			t.Fatalf("expected stopped exit code, got %q", code)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for queued item cancel exit")
	}

	close(first.release)
}

func TestControllerCancelRunningItemStopsWithStoppedExitCode(t *testing.T) {
	t.Parallel()

	controller := New("speed", 1, 0, time.Millisecond)
	stop := make(chan struct{})
	defer close(stop)

	go controller.Start(stop)

	item := newBlockingItem("running")
	controller.Push(item)

	select {
	case <-item.started:
	case <-time.After(time.Second):
		t.Fatalf("item did not start")
	}

	if ok := controller.Cancel("running"); !ok {
		t.Fatalf("expected running item to be cancelable")
	}

	close(item.release)

	select {
	case code := <-item.exitCode:
		if code != ExitCodeStopped {
			t.Fatalf("expected stopped exit code, got %q", code)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for stopped exit")
	}
}

type blockingItem struct {
	id       string
	started  chan struct{}
	release  chan struct{}
	exitCode chan ExitCode
	mu       sync.Mutex
	canceled bool
}

func newBlockingItem(id string) *blockingItem {
	return &blockingItem{
		id:       id,
		started:  make(chan struct{}),
		release:  make(chan struct{}),
		exitCode: make(chan ExitCode, 1),
	}
}

func (i *blockingItem) ID() string {
	return i.id
}

func (i *blockingItem) TaskName() string {
	return i.id
}

func (i *blockingItem) Count() int {
	return 1
}

func (i *blockingItem) Yield(idx int, c *Controller) {
	close(i.started)
	<-i.release
}

func (i *blockingItem) OnExit(exitCode ExitCode) {
	i.exitCode <- exitCode
}

func (i *blockingItem) Cancel() {
	i.mu.Lock()
	i.canceled = true
	i.mu.Unlock()
}

func (i *blockingItem) IsCanceled() bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.canceled
}
