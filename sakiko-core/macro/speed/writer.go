package speed

import "sync/atomic"

type writeCounter struct {
	total uint64
	last  uint64
}

func (w *writeCounter) Write(p []byte) (int, error) {
	atomic.AddUint64(&w.total, uint64(len(p)))
	return len(p), nil
}

func (w *writeCounter) Take() uint64 {
	total := atomic.LoadUint64(&w.total)
	last := atomic.LoadUint64(&w.last)
	delta := total - last
	atomic.StoreUint64(&w.last, total)
	return delta
}

func (w *writeCounter) Total() uint64 {
	return atomic.LoadUint64(&w.total)
}
