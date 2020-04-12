package handler

import (
	"context"

	"github.com/streadway/amqp"
)

// Handler handles incoming messages from the AMQP broker, which are supplied
// to the application using an amqp.Delivery object.
//
// An Handler is fallible, so it returns an error in case message handling
// failed.
//
// In that case, the handled message should be negatively-acknowledged,
// either requeued or rejected, depending on the error value.
type Handler interface {
	Handle(context.Context, amqp.Delivery) error
}

// Func is a function that implements the Handler interface, useful to specify
// Handlers inline as simple functions.
type Func func(context.Context, amqp.Delivery) error

// Handle calls the underlying function to handle the incoming message.
func (fn Func) Handle(ctx context.Context, delivery amqp.Delivery) error {
	return fn(ctx, delivery)
}
