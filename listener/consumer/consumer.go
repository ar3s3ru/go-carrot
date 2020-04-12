package consumer

import (
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

func (listener Listener) Listen(conn *amqp.Connection, ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		listener.queue,
		listener.queue,
		listener.autoAck,
		listener.exclusive,
		listener.noLocal,
		listener.noWait,
		listener.args,
	)
}

func (listener *Listener) addToTable(key string, value interface{}) {
	if listener.args == nil {
		listener.args = make(amqp.Table)
	}

	listener.args[key] = value
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
