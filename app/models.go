package app

import (
	"github.com/THE108/requestcounter/models/requestcount"
)

type models struct {
	requestCounter requestcount.IRequestCounter
}

func (this *Application) initModels() error {
	var counter requestcount.IRequestCounter
	if this.config.Persistent {
		counter = requestcount.NewPersistentRequestCounter(this.config.IntervalCount,
			this.config.IntervalDuration, this.config.Filename, this.config.PersistDuration)
	} else {
		counter = requestcount.NewRequestCounter(this.config.IntervalCount, this.config.IntervalDuration)
	}

	this.closer.AddCloser(counter)

	if err := counter.Run(); err != nil {
		return err
	}

	this.models = models{
		requestCounter: counter,
	}

	return nil
}
