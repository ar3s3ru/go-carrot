package listener

import (
	"fmt"
	"sync"

	"github.com/ar3s3ru/go-carrot/handler"
)

// Sink allows for listening from multiple Listeners, by maintaining
// an amqp.Delivery sink to which all the messages are sent.
//
// Returns nil if no listeners are supplied.
func Sink(listeners ...Listener) Listener {
	if len(listeners) == 0 {
		return nil
	}

	sinker := new(sinker)
	sinker.listeners = listeners

	return sinker
}

type sinker struct {
	sync.WaitGroup

	listeners []Listener
	closers   []Closer
	sink      chan error
}

func (sinker *sinker) Close() <-chan error {
	sinker.Add(len(sinker.closers))
	sinker.sink = make(chan error, len(sinker.closers))

	go func() {
		sinker.Wait()
		close(sinker.sink)
	}()

	for _, closer := range sinker.closers {
		go func(closer Closer) {
			defer sinker.Done()

			select {
			case err := <-closer.Close():
				sinker.sink <- err
			}
		}(closer)
	}

	return sinker.sink
}

func (sinker *sinker) Listen(conn Connection, ch Channel, h handler.Handler) (Closer, error) {
	if err := sinker.collectClosers(conn, ch, h); err != nil {
		return nil, err
	}

	return sinker, nil
}

func (sinker *sinker) collectClosers(conn Connection, ch Channel, h handler.Handler) error {
	if err := ch.Qos(len(sinker.listeners), 0, false); err != nil {
		return fmt.Errorf("listener.Sink: failed to set channel QoS, %w", err)
	}

	closers := make([]Closer, 0, len(sinker.listeners))

	for _, listener := range sinker.listeners {
		closer, err := listener.Listen(conn, ch, h)
		if err != nil {
			return fmt.Errorf("listener.Sink: failed to listen, %w", err)
		}

		closers = append(closers, closer)
	}

	sinker.closers = closers

	return nil
}
