package taskpoll

import (
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

type blockingItem struct {
	id      string
	started chan struct{}
	release chan struct{}
}

func newBlockingItem(id string) *blockingItem {
	return &blockingItem{
		id:      id,
		started: make(chan struct{}),
		release: make(chan struct{}),
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

func (i *blockingItem) OnExit(exitCode ExitCode) {}
