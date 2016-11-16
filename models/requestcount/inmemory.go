package requestcount

import (
	"sync"
	"time"

	"github.com/THE108/requestcounter/log"

	"golang.org/x/net/context"
)

type RequestCounter struct {
	mu               sync.Mutex
	counts           []uint64
	prevCountsSum    uint64
	index            int
	done             chan struct{}
	intervalDuration time.Duration
}

func NewRequestCounter(intervalCount int, intervalDuration time.Duration) *RequestCounter {
	return &RequestCounter{
		counts:           make([]uint64, intervalCount),
		done:             make(chan struct{}),
		intervalDuration: intervalDuration,
	}
}

func (rc *RequestCounter) Run() error {
	go rc.run()
	return nil
}

func (rc *RequestCounter) Close() error {
	rc.done <- struct{}{}
	return nil
}

func (rc *RequestCounter) run() {
	for {
		select {
		case <-time.After(rc.intervalDuration):
		case <-rc.done:
			return
		}

		rc.shift()
	}
}

func (rc *RequestCounter) shift() {
	rc.mu.Lock()

	rc.index++
	if rc.index >= len(rc.counts) {
		rc.index = 0
	}

	rc.counts[rc.index] = 0
	rc.prevCountsSum = 0
	for _, cnt := range rc.counts {
		rc.prevCountsSum += cnt
	}

	rc.mu.Unlock()
}

func (rc *RequestCounter) Get(ctx context.Context) *RequestCount {
	var count uint64
	rc.mu.Lock()
	rc.counts[rc.index]++
	count = rc.counts[rc.index] + rc.prevCountsSum
	rc.mu.Unlock()

	log.GetLoggerFromContext(ctx).Debugf("count: %d", count)

	return &RequestCount{
		Count: count,
	}
}
