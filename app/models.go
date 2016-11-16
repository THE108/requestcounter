package app

import (
	"github.com/THE108/requestcounter/models/requestcount"
)

type models struct {
	requestCounter requestcount.IRequestCounter
}

func (this *Application) initModels() error {
	counter := requestcount.NewRequestCounter(&requestcount.RequestCounterConfig{
		IntervalCount:    this.config.IntervalCount,
		IntervalDuration: this.config.IntervalDuration,
		Filename:         this.config.Filename,
		Persistent:       this.config.Persistent,
		PersistDuration:  this.config.PersistDuration,
		Logger:           this.logger,
	})

	this.closer.AddCloser(counter)

	if err := counter.Run(); err != nil {
		return err
	}

	this.models = models{
		requestCounter: counter,
	}

	return nil
}
