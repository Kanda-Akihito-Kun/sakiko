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
}

type Controller struct {
	name        string
	concurrency int
	interval    time.Duration
	idle        time.Duration
	queue       chan Item
	awaiting    atomic.Int64
	running     atomic.Int64
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
		queue:       make(chan Item, 256),
	}
}

func (c *Controller) Name() string {
	return c.name
}

func (c *Controller) Push(item Item) {
	c.awaiting.Add(1)
	c.queue <- item
}

func (c *Controller) AwaitingCount() int {
	return int(c.awaiting.Load())
}

func (c *Controller) RunningCount() int {
	return int(c.running.Load())
}

func (c *Controller) Start(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case item := <-c.queue:
			c.awaiting.Add(-1)
			c.running.Add(1)
			// Items are processed in queue order. The controller-level concurrency
			// is applied inside each item to bound node fan-out, instead of
			// allowing multiple heavyweight tasks to run at once.
			c.runItem(item)
		default:
			time.Sleep(c.idle)
		}
	}
}

func (c *Controller) runItem(item Item) {
	defer c.running.Add(-1)

	sem := make(chan struct{}, c.concurrency)
	var wg sync.WaitGroup
	for i := 0; i < item.Count(); i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() {
				<-sem
				if c.interval > 0 {
					time.Sleep(c.interval)
				}
			}()
			item.Yield(idx, c)
		}(i)
	}
	wg.Wait()
	item.OnExit(ExitCodeSuccess)
}
