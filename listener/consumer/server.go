package consumer

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/listener"

	"github.com/streadway/amqp"
)

// ErrAlreadyClosed is returned by the consumer when closing
// a previously-closed Listener.
var ErrAlreadyClosed = errors.New("consumer.Listener: already closed")

type server struct {
	conn listener.Connection
	ch   listener.Channel
	sink <-chan amqp.Delivery

	closeOnce *sync.Once
	close     chan error
	done      chan bool

	onError   func(amqp.Delivery, error)
	onSuccess func(amqp.Delivery)
}

func (srv *server) Close(ctx context.Context) error {
	err := ErrAlreadyClosed

	srv.closeOnce.Do(func() {
		err = srv.ch.Close()

		select {
		case <-srv.done:
		case <-ctx.Done():
			err = fmt.Errorf("consumer.Listener: failed to close server, %w", err)
		}

		srv.close <- err
		close(srv.close)
	})

	return err
}

func (srv *server) Closed() <-chan error {
	return srv.close
}

func (srv *server) serve(h handler.Handler) {
	for delivery := range srv.sink {
		switch err := h.Handle(context.Background(), delivery); err {
		case nil:
			srv.handleSuccess(delivery)
		default:
			srv.handleError(delivery, err)
		}
	}

	srv.done <- true
	close(srv.done)
}

// nolint:errcheck
func (srv *server) handleError(delivery amqp.Delivery, err error) {
	if srv.onError != nil {
		srv.onError(delivery, err)
	} else {
		delivery.Nack(false, true)
	}
}

// nolint:errcheck
func (srv *server) handleSuccess(delivery amqp.Delivery) {
	if srv.onSuccess != nil {
		srv.onSuccess(delivery)
	} else {
		delivery.Ack(false)
	}

}
