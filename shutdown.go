package carrot

import (
	"context"
	"os"
	"os/signal"
	"time"
)

var DefaultShutdownOptions = Shutdown{
	Timeout: 1 * time.Minute,
	Signals: []os.Signal{os.Interrupt},
	OnError: func(err error) {
		panic(err)
	},
}

type Shutdown struct {
	Timeout time.Duration
	Signals []os.Signal
	OnError func(err error)
}

func (shutdown *Shutdown) orDefault() Shutdown {
	options := DefaultShutdownOptions

	if shutdown == nil {
		return options
	}

	if timeout := shutdown.Timeout; timeout != 0 {
		options.Timeout = timeout
	}
	if signals := shutdown.Signals; len(signals) > 0 {
		options.Signals = signals
	}
	if onError := shutdown.OnError; onError != nil {
		options.OnError = onError
	}

	return options
}

func gracefulShutdown(closer Closer, options Shutdown) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, options.Signals...)

	<-c

	basectx := context.Background()
	ctx, cancel := context.WithTimeout(basectx, options.Timeout)
	defer cancel()

	err := closer.Close(ctx)
	if err != nil {
		options.OnError(err)
	}
}
