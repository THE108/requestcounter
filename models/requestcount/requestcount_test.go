package requestcount

import (
	"testing"
	"time"

	"github.com/THE108/requestcounter/utils/log"
	"github.com/THE108/requestcounter/utils/storage"

	"golang.org/x/net/context"
	. "gopkg.in/check.v1"
)

type RequestCounterSuite struct{}

var _ = Suite(&RequestCounterSuite{})

// Hook up gocheck into the "go test" runner.
func TestStart(t *testing.T) {
	TestingT(t)
}

func (suite *RequestCounterSuite) Test_Success(c *C) {
	devnull := log.NewDevNullLogger()
	ctx := log.SetLoggerToContext(context.Background(), devnull)
	fakeNow := time.Unix(0, 0)
	counter := &RequestCounter{
		done:             make(chan struct{}),
		intervalCount:    5,
		intervalDuration: 1,
		persistDuration:  1,
		logger:           devnull,
		now: func() time.Time {
			return fakeNow
		},
		storage: storage.NewInmemoryStorage(),
	}

	c.Assert(counter.Run(), IsNil)

	c.Assert(counter.Get(ctx).Count, Equals, uint64(1))
	c.Assert(counter.counts[0], Equals, uint64(0))
	c.Assert(counter.counts[1], Equals, uint64(fakeNow.UnixNano()))
	c.Assert(counter.counts[2], Equals, uint64(1))

	c.Assert(counter.Get(ctx).Count, Equals, uint64(2))
	c.Assert(counter.counts[0], Equals, uint64(0))
	c.Assert(counter.counts[1], Equals, uint64(fakeNow.UnixNano()))
	c.Assert(counter.counts[2], Equals, uint64(2))
}

func (suite *RequestCounterSuite) Test_Loop(c *C) {
	devnull := log.NewDevNullLogger()
	ctx := log.SetLoggerToContext(context.Background(), devnull)
	fakeNow := time.Unix(0, 0)
	counter := &RequestCounter{
		counts:           make([]uint64, 7),
		intervalCount:    5,
		intervalDuration: 1,
		persistDuration:  1,
		logger:           devnull,
		now: func() time.Time {
			return fakeNow
		},
		storage: storage.NewInmemoryStorage(),
	}

	for i := 0; i < 5; i++ {
		c.Assert(counter.Get(ctx).Count, Equals, uint64(i+1))
		c.Assert(counter.counts[0], Equals, uint64(i))
		c.Assert(counter.counts[1], Equals, uint64(fakeNow.UnixNano()))
		c.Assert(counter.counts[i+2], Equals, uint64(1))
		counter.shift(fakeNow)
	}
}

func (suite *RequestCounterSuite) Test_Loop2(c *C) {
	devnull := log.NewDevNullLogger()
	ctx := log.SetLoggerToContext(context.Background(), devnull)
	fakeNow := time.Unix(0, 0)
	counter := &RequestCounter{
		counts:           make([]uint64, 7),
		intervalCount:    5,
		intervalDuration: 1,
		persistDuration:  1,
		logger:           devnull,
		now: func() time.Time {
			return fakeNow
		},
		storage: storage.NewInmemoryStorage(),
	}

	for i := 0; i < 8; i++ {
		counter.Get(ctx)
		counter.shift(fakeNow)
	}

	counter.Get(ctx)

	c.Assert(counter.counts[0], Equals, uint64(3))
	c.Assert(counter.counts[1], Equals, uint64(fakeNow.UnixNano()))
	for i := 2; i < 7; i++ {
		c.Assert(counter.counts[i], Equals, uint64(1), Commentf("i: %d", i))
	}
}

func (suite *RequestCounterSuite) Test_Restart_ClearAll(c *C) {
	devnull := log.NewDevNullLogger()
	ctx := log.SetLoggerToContext(context.Background(), devnull)
	fakeNow := time.Unix(0, 0)
	counter := &RequestCounter{
		counts:           make([]uint64, 7),
		intervalCount:    5,
		intervalDuration: 1,
		persistDuration:  1,
		logger:           devnull,
		now: func() time.Time {
			return fakeNow
		},
		storage: storage.NewInmemoryStorage(),
	}

	for i := 0; i < 8; i++ {
		counter.Get(ctx)
		counter.shift(time.Unix(0, 0))
	}
	counter.Get(ctx)
	fakeNow = time.Unix(0, 10)
	counter.clearOutdated()
	counter.calculatePrevCountSum()

	c.Assert(counter.counts, DeepEquals, []uint64{3, 0, 0, 0, 0, 0, 0})
	c.Assert(counter.prevCountsSum, Equals, uint64(0))
}

func (suite *RequestCounterSuite) Test_Restart_ClearSome(c *C) {
	devnull := log.NewDevNullLogger()
	ctx := log.SetLoggerToContext(context.Background(), devnull)
	fakeNow := time.Unix(0, 0)
	counter := &RequestCounter{
		counts:           make([]uint64, 7),
		intervalCount:    5,
		intervalDuration: 1,
		persistDuration:  1,
		logger:           devnull,
		now: func() time.Time {
			return fakeNow
		},
		storage: storage.NewInmemoryStorage(),
	}

	for i := 0; i < 8; i++ {
		counter.Get(ctx)
		counter.shift(time.Unix(0, 0))
	}
	counter.Get(ctx)
	fakeNow = time.Unix(0, 2)
	counter.clearOutdated()
	counter.calculatePrevCountSum()

	c.Assert(counter.counts, DeepEquals, []uint64{3, 0, 0, 0, 1, 1, 0})
	c.Assert(counter.prevCountsSum, Equals, uint64(2))
}
