package speed

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"sakiko.local/sakiko-core/interfaces"
	"sakiko.local/sakiko-core/netx"
)

type Macro struct {
	AvgSpeed uint64
	MaxSpeed uint64
	Total    uint64
	Speeds   []uint64
}

const (
	switchCooldown = time.Second
)

func (m *Macro) Type() interfaces.MacroType {
	return interfaces.MacroSpeed
}

func (m *Macro) Run(proxy interfaces.Vendor, task *interfaces.Task) error {
	cfg := task.Config.Normalize()
	time.Sleep(switchCooldown)

	samples := make([]uint64, cfg.DownloadDuration)

	counters := make([]*writeCounter, 0, cfg.DownloadThreading)
	cancels := make([]context.CancelFunc, 0, cfg.DownloadThreading)
	var ready sync.WaitGroup

	for i := 0; i < int(cfg.DownloadThreading); i++ {
		ready.Add(1)
		counter := &writeCounter{}
		cancel := singleThread(proxy, cfg.DownloadURL, cfg.DownloadDuration, counter, &ready)
		counters = append(counters, counter)
		cancels = append(cancels, cancel)
	}
	ready.Wait()

	for i := 0; i < int(cfg.DownloadDuration); i++ {
		time.Sleep(time.Second)
		var speedThisSecond uint64
		for _, counter := range counters {
			speedThisSecond += counter.Take()
		}
		samples[i] = speedThisSecond
	}
	for _, cancel := range cancels {
		cancel()
	}

	m.Speeds = dropExtremes(samples)
	for _, speedThisSecond := range m.Speeds {
		m.Total += speedThisSecond
		if speedThisSecond > m.MaxSpeed {
			m.MaxSpeed = speedThisSecond
		}
	}
	if len(m.Speeds) > 0 {
		m.AvgSpeed = m.Total / uint64(len(m.Speeds))
	}
	if m.Total == 0 {
		return fmt.Errorf("speed test failed, with total in 0")
	}
	return nil
}

func singleThread(proxy interfaces.Vendor, rawURL string, duration int64, counter *writeCounter, ready *sync.WaitGroup) context.CancelFunc {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration+1)*time.Second)
	go func() {
		ready.Done()
		for ctx.Err() == nil {
			resp, err := netx.RequestUnsafe(ctx, proxy, interfaces.RequestOptions{
				URL:     rawURL,
				Method:  "GET",
				Network: interfaces.ROptionsTCP,
			})
			if err != nil {
				time.Sleep(150 * time.Millisecond)
				continue
			}
			_, _ = io.Copy(io.Discard, io.TeeReader(resp.Body, counter))
			resp.Body.Close()
		}
	}()
	return cancel
}

func dropExtremes(samples []uint64) []uint64 {
	if len(samples) <= 2 {
		return append([]uint64{}, samples...)
	}

	minIdx := 0
	maxIdx := 0
	for i := 1; i < len(samples); i++ {
		if samples[i] < samples[minIdx] {
			minIdx = i
		}
		if samples[i] >= samples[maxIdx] {
			maxIdx = i
		}
	}

	filtered := make([]uint64, 0, len(samples)-2)
	for i, sample := range samples {
		if i == minIdx || i == maxIdx {
			continue
		}
		filtered = append(filtered, sample)
	}
	return filtered
}
