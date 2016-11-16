package requestcount

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"
	"unsafe"

	"github.com/THE108/requestcounter/log"

	"github.com/edsrzf/mmap-go"
	"golang.org/x/net/context"
)

type PersistentRequestCounter struct {
	mu               sync.Mutex
	counts           []uint64
	prevCountsSum    uint64
	done             chan struct{}
	intervalCount    int
	intervalDuration time.Duration
	filename         string
	file             *os.File
	mmaped           mmap.MMap
	closed           bool
	persistDuration  time.Duration
	wg               sync.WaitGroup
	logger           log.ILogger
}

func NewPersistentRequestCounter(intervalCount int, intervalDuration time.Duration, filename string,
	persistDuration time.Duration) *PersistentRequestCounter {
	return &PersistentRequestCounter{
		done:             make(chan struct{}),
		intervalCount:    intervalCount,
		intervalDuration: intervalDuration,
		filename:         filename,
		persistDuration:  persistDuration,
		logger:           log.New(os.Stderr, "ASYNC", log.DEBUG),
	}
}

func (prc *PersistentRequestCounter) Run() error {
	var err error
	prc.file, err = os.OpenFile(prc.filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error open file: %s", err.Error())
	}

	// counts[0] - current index
	// counts[1] - current index init timestamp
	length := prc.intervalCount + 2

	// prealloc intervalCount + 2 uint64 values
	offset := int64(unsafe.Sizeof(prc.prevCountsSum)) * int64(length)
	if _, err = prc.file.Seek(offset, 0); err != nil {
		prc.closeFile()
		return fmt.Errorf("error seek file: %s", err.Error())
	}

	// write byte to change size of the file
	if _, err = prc.file.Write([]byte{0xFF}); err != nil {
		prc.closeFile()
		return fmt.Errorf("error write file: %s", err.Error())
	}

	// map file in memory
	prc.mmaped, err = mmap.Map(prc.file, mmap.RDWR, 0)
	if err != nil {
		prc.closeFile()
		return fmt.Errorf("error map file: %s", err.Error())
	}

	// cast mapped []byte to []uint64
	header := (*reflect.SliceHeader)(unsafe.Pointer(&prc.mmaped))
	header.Len, header.Cap = length, length
	prc.counts = *(*[]uint64)(unsafe.Pointer(header))

	prc.clearOutdated()
	prc.calculatePrevCountSum()

	prc.wg.Add(2)
	go prc.runShift()
	go prc.runPersist()

	return nil
}

func (prc *PersistentRequestCounter) closeFile() {
	prc.logger.ErrorIfNotNil("error close file:", prc.file.Close())
}

func (prc *PersistentRequestCounter) Close() error {
	close(prc.done)
	prc.wg.Wait()

	prc.logger.Debug("after wait")

	prc.mu.Lock()
	defer prc.mu.Unlock()

	prc.logger.Debug("after mutex")

	prc.closed = true

	if err := prc.mmaped.Flush(); err != nil {
		return err
	}

	prc.logger.Debug("after flush")

	if err := prc.mmaped.Unmap(); err != nil {
		return err
	}

	prc.logger.Debug("after unmap")

	if err := prc.file.Close(); err != nil {
		return err
	}

	prc.logger.Debug("after close file")

	return nil
}

func (prc *PersistentRequestCounter) Get(ctx context.Context) *RequestCount {
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

func (prc *PersistentRequestCounter) clearOutdated() {
	prevTimestamp := time.Unix(0, int64(prc.counts[1]))
	now := time.Now()
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

func (prc *PersistentRequestCounter) calculatePrevCountSum() {
	prc.prevCountsSum = 0
	for _, cnt := range prc.counts[2:] {
		prc.prevCountsSum += cnt
	}
}

func (prc *PersistentRequestCounter) shift() {
	prc.mu.Lock()

	// counts[0] - current index
	prc.counts[0]++
	if int(prc.counts[0]) >= prc.intervalCount {
		prc.counts[0] = 0
	}

	// set timestamp milliseconds
	prc.counts[1] = uint64(time.Now().UnixNano())

	// set current request count to 0
	prc.counts[int(prc.counts[0])+2] = 0

	prc.calculatePrevCountSum()

	prc.mu.Unlock()
}

func (prc *PersistentRequestCounter) runShift() {
	defer prc.wg.Done()
	for {
		select {
		case <-time.After(prc.intervalDuration):
		case <-prc.done:
			prc.logger.Debug("runShift is done")
			return
		}

		prc.shift()
	}
}

func (prc *PersistentRequestCounter) persist() {
	prc.logger.ErrorIfNotNil("error flush mmaped file:", prc.mmaped.Flush())
}

func (prc *PersistentRequestCounter) runPersist() {
	defer prc.wg.Done()
	for {
		select {
		case <-time.After(prc.persistDuration):
		case <-prc.done:
			prc.logger.Debug("runPersist is done")
			return
		}

		prc.persist()
	}
}
