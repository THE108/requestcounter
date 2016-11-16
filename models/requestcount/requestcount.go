package requestcount

import (
	"sync"
	"time"

	"github.com/THE108/requestcounter/utils/log"
	"github.com/THE108/requestcounter/utils/storage"
	utime "github.com/THE108/requestcounter/utils/time"

	"golang.org/x/net/context"
)

type RequestCount struct {
	Count uint64 `json:"count"`
}

type IRequestCounter interface {
	Get(ctx context.Context) *RequestCount
	Run() error
	Close() error
}

type IStorage interface {
	Open(filename string, length int) ([]uint64, error)
	Close() error
	Flush() error
}

type RequestCounterConfig struct {
	IntervalCount    int
	IntervalDuration time.Duration
	Filename         string
	Persistent       bool
	PersistDuration  time.Duration
	Logger           log.ILogger
}

type RequestCounter struct {
	mu               sync.Mutex
	counts           []uint64
	prevCountsSum    uint64
	intervalCount    int
	intervalDuration time.Duration
	filename         string
	persistent       bool
	closed           bool
	persistDuration  time.Duration
	done             chan struct{}
	wg               sync.WaitGroup
	logger           log.ILogger
	timeManager      utime.ITime
	storage          IStorage
}

func NewRequestCounter(cfg *RequestCounterConfig) *RequestCounter {
	var st IStorage
	if cfg.Persistent {
		st = storage.NewPersistentStorage()
	} else {
		st = storage.NewInmemoryStorage()
	}

	return &RequestCounter{
		done:             make(chan struct{}),
		intervalCount:    cfg.IntervalCount,
		intervalDuration: cfg.IntervalDuration,
		filename:         cfg.Filename,
		persistent:       cfg.Persistent,
		persistDuration:  cfg.PersistDuration,
		logger:           cfg.Logger,
		timeManager:      utime.NewRealTime(),
		storage:          st,
	}
}

func (prc *RequestCounter) Run() error {
	// prealloc intervalCount + 2 uint64 values
	// counts[0] - current index
	// counts[1] - current index init timestamp
	length := prc.intervalCount + 2

	var err error
	prc.counts, err = prc.storage.Open(prc.filename, length)
	if err != nil {
		return err
	}

	prc.clearOutdated()
	prc.calculatePrevCountSum()

	prc.wg.Add(1)
	go prc.runShift()

	if prc.persistent {
		prc.wg.Add(1)
		go prc.runPersist()
	}

	return nil
}

func (prc *RequestCounter) Close() error {
	close(prc.done)
	prc.wg.Wait()

	prc.mu.Lock()
	defer prc.mu.Unlock()

	prc.closed = true

	if err := prc.storage.Close(); err != nil {
		return err
	}

	return nil
}

func (prc *RequestCounter) Get(ctx context.Context) *RequestCount {
	var count uint64
	prc.mu.Lock()
	if prc.closed {
		prc.mu.Unlock()
		return nil
	}
	index := int(prc.counts[0]) + 2
	prc.counts[index]++
	count = prc.counts[index] + prc.prevCountsSum
	prc.mu.Unlock()

	log.GetLoggerFromContext(ctx).Debugf("count: %d", count)

	return &RequestCount{
		Count: count,
	}
}

func (prc *RequestCounter) clearOutdated() {
	prevTimestamp := time.Unix(0, int64(prc.counts[1]))
	now := prc.timeManager.Now()
	startIndex := int(prc.counts[0]) + 1
	for i := 0; i < prc.intervalCount; i++ {
		t := prevTimestamp.Add(prc.intervalDuration * time.Duration(i))
		if t.After(now) {
			continue
		}

		index := startIndex + i
		if index >= prc.intervalCount {
			index = 0
		}

		prc.counts[index+2] = 0
	}
}

func (prc *RequestCounter) calculatePrevCountSum() {
	prc.prevCountsSum = 0
	for _, cnt := range prc.counts[2:] {
		prc.prevCountsSum += cnt
	}
}

func (prc *RequestCounter) shift(now time.Time) {
	prc.mu.Lock()

	// counts[0] - current index
	prc.counts[0]++
	if int(prc.counts[0]) >= prc.intervalCount {
		prc.counts[0] = 0
	}

	// set timestamp milliseconds
	prc.counts[1] = uint64(now.UnixNano())

	// set current request count to 0
	prc.counts[int(prc.counts[0])+2] = 0

	prc.calculatePrevCountSum()

	prc.mu.Unlock()
}

func (prc *RequestCounter) runShift() {
	defer prc.wg.Done()
	var now time.Time
	for {
		select {
		case now = <-prc.timeManager.After(prc.intervalDuration):
		case <-prc.done:
			prc.logger.Debug("runShift is done")
			return
		}

		prc.shift(now)
	}
}

func (prc *RequestCounter) persist() {
	prc.logger.ErrorIfNotNil("error flush mmaped file:", prc.storage.Flush())
}

func (prc *RequestCounter) runPersist() {
	defer prc.wg.Done()
	for {
		select {
		case <-prc.timeManager.After(prc.persistDuration):
		case <-prc.done:
			prc.logger.Debug("runPersist is done")
			return
		}

		prc.persist()
	}
}
