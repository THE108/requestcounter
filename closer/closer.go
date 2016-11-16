package closer

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/THE108/requestcounter/log"
)

type ICloser interface {
	Close() error
}

type Closer struct {
	mu      sync.Mutex
	closers []ICloser
}

func NewCloser() *Closer {
	return &Closer{}
}

func (c *Closer) AddCloser(closer ICloser) {
	c.mu.Lock()
	c.closers = append(c.closers, closer)
	c.mu.Unlock()
}

func (c *Closer) Run() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)

	logger := log.New(os.Stderr, "ASYNC", log.INFO)

	logger.Info("start waiting for signals")

	sig := <-ch

	logger.Info("got signal ", sig)
	logger.ErrorIfNotNil("error on closing", c.close())
}

func (c *Closer) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, closer := range c.closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}
