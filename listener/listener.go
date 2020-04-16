package listener

import (
	"context"
	"io"

	"github.com/ar3s3ru/go-carrot/handler"

	"github.com/streadway/amqp"
)

// Connection is the connection interface the Listener uses to listen and serve
// incoming messages.
type Connection interface {
	Channel() (*amqp.Channel, error)
}

// Channel is the channel interface the Listener uses to listen and serve
// incoming messages.
type Channel interface {
	io.Closer

	Qos(prefetchCount, prefetchSize int, global bool) error
	Consume(queue, name string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

// Closer is used to stop listening from a specific Listener.
type Closer interface {
	Close(context.Context) error
	Closed() <-chan error
}

// Listener listens for incoming messages using an amqp.Connection or
// amqp.Channel provided, and calls the specified message handler
// to handle all incoming messages.
//
// Returns a Closer interface to stop listening to incoming messages.
type Listener interface {
	Listen(Connection, Channel, handler.Handler) (Closer, error)
}

// Func is an inline function that implements the Listener interface.
// Useful for mocking or stateless Listener implementations.
type Func func(Connection, Channel, handler.Handler) (Closer, error)

// Listen executes the inline function.
func (fn Func) Listen(conn Connection, ch Channel, h handler.Handler) (Closer, error) {
	return fn(conn, ch, h)
}
