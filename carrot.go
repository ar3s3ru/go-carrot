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

var ErrNoConnection = errors.New("carrot: no connection provided")
var ErrNoHandler = errors.New("carrot: no handler specified")
var ErrNoBinder = errors.New("carrot: no consumer binder specified")

type Runner struct {
	conn     *amqp.Connection
	declarer topology.Declarer
	handler  handler.Handler
	binder   binder.Binder
}

func (runner Runner) Run() error {
	if runner.conn == nil {
		return ErrNoConnection
	}

	if runner.handler == nil {
		return ErrNoHandler
	}

	if runner.binder == nil {
		return ErrNoBinder
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

type Option func(*Runner)

func WithTopology(declarer topology.Declarer) Option {
	return func(runner *Runner) { runner.declarer = declarer }
}

func WithHandler(handler handler.Handler) Option {
	return func(runner *Runner) { runner.handler = handler }
}

func WithBinder(binder binder.Binder) Option {
	return func(runner *Runner) { runner.binder = binder }
}
