package consumer

import (
	"context"
	"fmt"

	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/listener"

	"github.com/streadway/amqp"
)

type Listener struct {
	queue       string
	title       string
	description string

	autoAck   bool
	exclusive bool
	noLocal   bool
	noWait    bool
	args      amqp.Table
}

func (l Listener) Listen(conn listener.Connection, ch listener.Channel, h handler.Handler) (listener.Closer, error) {
	delivery, err := ch.Consume(l.queue, l.queue, l.autoAck, l.exclusive, l.noLocal, l.noWait, l.args)
	if err != nil {
		return nil, fmt.Errorf("consumer.Listener: failed to start consuming messages, %w", err)
	}

	runner := runner{
		conn:  conn,
		ch:    ch,
		sink:  delivery,
		close: make(chan error),
	}

	go runner.serve(h)

	return &runner, nil
}

func (l *Listener) addToTable(key string, value interface{}) {
	if l.args == nil {
		l.args = make(amqp.Table)
	}

	l.args[key] = value
}

type runner struct {
	conn  listener.Connection
	ch    listener.Channel
	sink  <-chan amqp.Delivery
	close chan error
}

func (r *runner) Close() <-chan error {
	go func() {
		defer close(r.close)
		r.close <- r.ch.Close()
	}()

	return r.close
}

func (r *runner) serve(h handler.Handler) {
	for delivery := range r.sink {
		h.Handle(context.Background(), delivery)
	}
}

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

type Option func(*Listener)

func Title(title string) Option {
	return func(listener *Listener) { listener.title = title }
}

func Description(description string) Option {
	return func(listener *Listener) { listener.description = description }
}

func AutoAck(listener *Listener) { listener.autoAck = true }

func Exclusive(listener *Listener) { listener.exclusive = true }

func NoLocal(listener *Listener) { listener.noLocal = true }

func NoWait(listener *Listener) { listener.noWait = true }

func Arguments(args amqp.Table) Option {
	return func(listener *Listener) {
		for key, value := range args {
			listener.addToTable(key, value)
		}
	}
}
