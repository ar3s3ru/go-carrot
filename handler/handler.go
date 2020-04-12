package handler

import (
	"context"

	"github.com/streadway/amqp"
)

type Handler interface {
	Handle(context.Context, amqp.Delivery) error
}

type Func func(context.Context, amqp.Delivery) error

func (fn Func) Handle(ctx context.Context, delivery amqp.Delivery) error {
	return fn(ctx, delivery)
}
