package carrot

import (
	"context"

	"github.com/streadway/amqp"
)

type Handler interface {
	Handle(context.Context, amqp.Delivery) error
}

type HandlerFunc func(context.Context, amqp.Delivery) error

func (fn HandlerFunc) Handle(ctx context.Context, delivery amqp.Delivery) error {
	return fn(ctx, delivery)
}
