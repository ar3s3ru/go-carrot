package carrot

import (
	"context"
	"errors"
	"fmt"

	"github.com/ar3s3ru/go-carrot/binder"
	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/topology"

	"github.com/streadway/amqp"
)

// ErrNoConnection is returned by Runner.Run when no valid AMQP connection
// has been specified.
var ErrNoConnection = errors.New("carrot: no connection provided")

// ErrNoHandler is returned by Runner.Run when no handler has been specified,
// so that Runner.Run can't handle any incoming messages.
var ErrNoHandler = errors.New("carrot: no handler specified")

// ErrNoBinder is returned by Runner.Run when no delivery binder
// has been specified, so that Runner.Run can't receive any messages
// from the AMQP broker.
var ErrNoBinder = errors.New("carrot: no consumer binder specified")

// Runner instruments all the different parts of the go-carrot library,
// provided with a valid AMQP connection.
type Runner struct {
	conn     *amqp.Connection
	declarer topology.Declarer
	handler  handler.Handler
	binder   binder.Binder
}

// Run starts all the different parts of the Runner instrumentator,
// in the following order: topology declaration, delivery binder and messages listener.
//
// Message listener uses the sink channel coming from the delivery binder,
// and spawns a separate worker goroutine to run the message handler
// specified during configuration with the new amqp.Delivery received.
//
// An error is returned if the supplied parameters during configuration are not
// valid, or if something happened on the AMQP connection.
func (runner Runner) Run() error {
	if runner.conn == nil {
		return ErrNoConnection
	}

	ch, err := runner.conn.Channel()
	if err != nil {
		return fmt.Errorf("carrot: failed to create channel from connection, %w", err)
	}

	if runner.declarer != nil {
		if err := runner.declarer.Declare(ch); err != nil {
			return fmt.Errorf("carrot: failed to declare topology, %w", err)
		}
	}

	// No handler nor delivery binder is an acceptable scenario: it means
	// the user is not leveraging carrot for message consumption.
	if runner.handler == nil && runner.binder == nil {
		return nil
	}

	if runner.handler == nil {
		return ErrNoHandler
	}

	if runner.binder == nil {
		return ErrNoBinder
	}

	rx, err := runner.binder.Bind(runner.conn, ch)
	if err != nil {
		return fmt.Errorf("carrot: failed to bind, %w", err)
	}

	go func(rx <-chan amqp.Delivery) {
		for delivery := range rx {
			go func(delivery amqp.Delivery) { runner.handler.Handle(context.Background(), delivery) }(delivery)
		}
	}(rx)

	return nil
}

// From creates a new Runner instance, given an AMQP connection and options.
//
// Required options are WithBinder, to bind a channel to an amqp.Delivery sink
// and start receiving messages, and WithHandler, to handle all the incoming
// messages.
func From(conn *amqp.Connection, options ...Option) Runner {
	runner := Runner{conn: conn}

	for _, option := range options {
		if option == nil {
			continue
		}

		option(&runner)
	}

	return runner
}

// Option represents an additional argument for the Runner factory method.
type Option func(*Runner)

// WithTopology adds a topology declaration step to the new Runner instance.
func WithTopology(declarer topology.Declarer) Option {
	return func(runner *Runner) { runner.declarer = declarer }
}

// WithHandler specifies the component in charge of handling incoming messages
// for the new Runner instance.
func WithHandler(handler handler.Handler) Option {
	return func(runner *Runner) { runner.handler = handler }
}

// WithBinder specifies the component in charge of start listening messages
// coming from the AMQP broker.
func WithBinder(binder binder.Binder) Option {
	return func(runner *Runner) { runner.binder = binder }
}
