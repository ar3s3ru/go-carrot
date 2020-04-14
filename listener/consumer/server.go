package consumer

import (
	"context"

	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/listener"

	"github.com/streadway/amqp"
)

type server struct {
	conn  listener.Connection
	ch    listener.Channel
	sink  <-chan amqp.Delivery
	close chan error

	onError   func(amqp.Delivery, error)
	onSuccess func(amqp.Delivery)
}

func (srv *server) Close() <-chan error {
	go func() {
		defer close(srv.close)
		srv.close <- srv.ch.Close()
	}()

	return srv.close
}

func (srv server) serve(h handler.Handler) {
	for delivery := range srv.sink {
		switch err := h.Handle(context.Background(), delivery); err {
		case nil:
			srv.handleSuccess(delivery)
		default:
			srv.handleError(delivery, err)
		}
	}
}

// nolint:errcheck
func (srv server) handleError(delivery amqp.Delivery, err error) {
	if srv.onError != nil {
		srv.onError(delivery, err)
	} else {
		delivery.Nack(false, true)
	}
}

// nolint:errcheck
func (srv server) handleSuccess(delivery amqp.Delivery) {
	if srv.onSuccess != nil {
		srv.onSuccess(delivery)
	} else {
		delivery.Ack(false)
	}

}
