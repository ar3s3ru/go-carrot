package listener

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ar3s3ru/go-carrot/handler"

	"golang.org/x/sync/errgroup"
)

// ErrSinkAlreadyClosed is returned by listener.Sink when attempting
// to close the returned Closer more than once.
var ErrSinkAlreadyClosed = errors.New("listener.Sink: already closed")

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
	sinker.sink = make(chan error, 1)

	return sinker
}

type sinker struct {
	sync.WaitGroup

	listeners []Listener
	closers   []Closer
	sink      chan error
	closeOnce sync.Once
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

func (sinker *sinker) Close(ctx context.Context) error {
	err := ErrSinkAlreadyClosed

	sinker.closeOnce.Do(func() {
		g, ctx := errgroup.WithContext(ctx)

		for _, closer := range sinker.closers {
			closer := closer
			g.Go(func() error { return closer.Close(ctx) })
		}

		err = g.Wait()

		sinker.sink <- err
		close(sinker.sink)
	})

	return err
}

func (sinker *sinker) Closed() <-chan error {
	return sinker.sink
}
