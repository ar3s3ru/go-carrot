package consumer

import (
	"fmt"

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

	srv := server{
		conn:  conn,
		ch:    ch,
		sink:  delivery,
		close: make(chan error),
	}

	go srv.serve(h)

	return &srv, nil
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
