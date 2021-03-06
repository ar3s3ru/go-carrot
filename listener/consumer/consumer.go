package consumer

import (
	"fmt"
	"sync"

	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/listener"

	"github.com/streadway/amqp"
)

// Listener allows to listen for incoming messages from a consumer, which
// is attached to one or more queues.
//
// Use Listen function to create a new Listener instance.
type Listener struct {
	server

	queue       string
	title       string
	description string

	autoAck   bool
	exclusive bool
	noLocal   bool
	noWait    bool
	args      amqp.Table
}

// Listen starts listening to incoming messages on the specified queue (or queues)
// using the provided Connection or Channel, and serving them by using
// the provided Handler.
//
// An error is returned if the Listener is unable to start listening
// on the provided Channel.
func (l Listener) Listen(conn listener.Connection, ch listener.Channel, h handler.Handler) (listener.Closer, error) {
	delivery, err := ch.Consume(l.queue, l.queue, l.autoAck, l.exclusive, l.noLocal, l.noWait, l.args)
	if err != nil {
		return nil, fmt.Errorf("consumer.Listener: failed to start consuming messages, %w", err)
	}

	l.server.conn = conn
	l.server.ch = ch
	l.server.sink = delivery
	l.server.closeOnce = new(sync.Once)
	l.server.done = make(chan bool)
	// Needs buffer, in case user of the library doesn't listen to the close channel.
	l.server.close = make(chan error, 1)

	go l.serve(h)

	return &l, nil
}

func (l *Listener) addToTable(key string, value interface{}) {
	if l.args == nil {
		l.args = make(amqp.Table)
	}

	l.args[key] = value
}

// Listen returns a new Listener able to listen and serve incoming messages
// on one or more AMQP queues.
func Listen(queue string, options ...Option) Listener {
	listener := Listener{queue: queue}

	for _, option := range options {
		if option == nil {
			continue
		}

		option(&listener)
	}

	return listener
}

// Option is an optional functionality that can be added to the Listener
// that is being initialized by the Listen factory method.
type Option func(*Listener)

// Title adds a title for the consumer.
func Title(title string) Option {
	return func(listener *Listener) { listener.title = title }
}

// Description adds a description for the consumer.
func Description(description string) Option {
	return func(listener *Listener) { listener.description = description }
}

// AutoAck automatically acknowledges a message as soon as the AMQP broker delivers
// it to the Consumer.
//
// This might lead to message loss scenarios in case the consumer gets abruptly closed.
func AutoAck(listener *Listener) { listener.autoAck = true }

// Exclusive ensures that this consumer is the only one consuming messages
// for the specified queue.
func Exclusive(listener *Listener) { listener.exclusive = true }

// NoWait does not wait for the AMQP broker to confirm the "consume" request
// and immediately begin deliveries.
func NoWait(listener *Listener) { listener.noWait = true }

// Arguments specifies optional arguments to be supplied during queue declaration.
//
// Multiple calls of this option are supported.
func Arguments(args amqp.Table) Option {
	return func(listener *Listener) {
		for key, value := range args {
			listener.addToTable(key, value)
		}
	}
}

// OnSuccess specifies the callback function to execute when the message handler
// successfully processed the message (i.e. failed without an error).
//
// If not specified, the Server will call amqp.Delivery.Ack(false).
func OnSuccess(fn func(amqp.Delivery)) Option {
	return func(listener *Listener) { listener.onSuccess = fn }
}

// OnError specifies the callback function to execute when the message handler
// fails with an error, providing the amqp.Delivery and the error returned
// by the handler itself.
//
// If not specified, the Server will call amqp.Delivery.Nack(false, true).
func OnError(fn func(amqp.Delivery, error)) Option {
	return func(listener *Listener) { listener.onError = fn }
}
