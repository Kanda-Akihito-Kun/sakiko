package taskpoll

import (
	"sync"
	"sync/atomic"
	"time"
)

type ExitCode string

const (
	ExitCodeSuccess ExitCode = "success"
	ExitCodeStopped ExitCode = "stopped"
)

type Item interface {
	ID() string
	TaskName() string
	Count() int
	Yield(idx int, c *Controller)
	OnExit(exitCode ExitCode)
	Cancel()
	IsCanceled() bool
}

type Controller struct {
	name        string
	concurrency int
	interval    time.Duration
	idle        time.Duration
	awaiting    atomic.Int64
	running     atomic.Int64

	lock        sync.Mutex
	queue       []Item
	queued      map[string]Item
	runningItem Item
}

func New(name string, concurrency uint, interval time.Duration, idle time.Duration) *Controller {
	if concurrency == 0 {
		concurrency = 1
	}
	if idle <= 0 {
		idle = 100 * time.Millisecond
	}
	return &Controller{
		name:        name,
		concurrency: int(concurrency),
		interval:    interval,
		idle:        idle,
		queue:       make([]Item, 0, 256),
		queued:      map[string]Item{},
	}
}

func (c *Controller) Name() string {
	return c.name
}

func (c *Controller) Push(item Item) {
	c.lock.Lock()
	c.queue = append(c.queue, item)
	c.queued[item.ID()] = item
	c.lock.Unlock()
	c.awaiting.Add(1)
}

func (c *Controller) AwaitingCount() int {
	return int(c.awaiting.Load())
}

func (c *Controller) RunningCount() int {
	return int(c.running.Load())
}

func (c *Controller) Cancel(taskID string) bool {
	c.lock.Lock()
	if item, ok := c.queued[taskID]; ok {
		for i, queued := range c.queue {
			if queued.ID() != taskID {
				continue
			}
			c.queue = append(c.queue[:i], c.queue[i+1:]...)
			break
		}
		delete(c.queued, taskID)
		c.awaiting.Add(-1)
		c.lock.Unlock()
		item.Cancel()
		item.OnExit(ExitCodeStopped)
		return true
	}

	running := c.runningItem
	if running != nil && running.ID() == taskID {
		running.Cancel()
		c.lock.Unlock()
		return true
	}
	c.lock.Unlock()
	return false
}

func (c *Controller) Start(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		item := c.popNext()
		if item == nil {
			time.Sleep(c.idle)
			continue
		}

		if item.IsCanceled() {
			item.OnExit(ExitCodeStopped)
			continue
		}

		c.running.Add(1)
		c.lock.Lock()
		c.runningItem = item
		c.lock.Unlock()
		// Items are processed in queue order. The controller-level concurrency
		// is applied inside each item to bound node fan-out, instead of
		// allowing multiple heavyweight tasks to run at once.
		c.runItem(item)
	}
}

func (c *Controller) runItem(item Item) {
	defer func() {
		c.lock.Lock()
		if c.runningItem != nil && c.runningItem.ID() == item.ID() {
			c.runningItem = nil
		}
		c.lock.Unlock()
		c.running.Add(-1)
	}()

	sem := make(chan struct{}, c.concurrency)
	var wg sync.WaitGroup
	exitCode := ExitCodeSuccess
	for i := 0; i < item.Count(); i++ {
		if item.IsCanceled() {
			exitCode = ExitCodeStopped
			break
		}

		for {
			if item.IsCanceled() {
				exitCode = ExitCodeStopped
				break
			}
			select {
			case sem <- struct{}{}:
				goto launch
			default:
				time.Sleep(c.idle)
			}
		}
		if exitCode == ExitCodeStopped {
			break
		}

	launch:
		if item.IsCanceled() {
			<-sem
			exitCode = ExitCodeStopped
			break
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() {
				<-sem
				if c.interval > 0 {
					time.Sleep(c.interval)
				}
			}()
			if item.IsCanceled() {
				return
			}
			item.Yield(idx, c)
		}(i)
	}
	wg.Wait()
	if item.IsCanceled() {
		exitCode = ExitCodeStopped
	}
	item.OnExit(exitCode)
}

func (c *Controller) popNext() Item {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.queue) == 0 {
		return nil
	}
	item := c.queue[0]
	c.queue = c.queue[1:]
	delete(c.queued, item.ID())
	c.awaiting.Add(-1)
	return item
}
