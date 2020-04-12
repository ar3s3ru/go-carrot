package consumer

import (
	"github.com/streadway/amqp"
)

type Binder struct {
	queue       string
	title       string
	description string

	autoAck   bool
	exclusive bool
	noLocal   bool
	noWait    bool
	args      amqp.Table

	useNewChannel bool
}

func (binder Binder) Bind(conn *amqp.Connection, ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	var err error

	if binder.useNewChannel {
		if ch, err = conn.Channel(); err != nil {
			return nil, err
		}
	}

	return ch.Consume(
		binder.queue,
		binder.queue,
		binder.autoAck,
		binder.exclusive,
		binder.noLocal,
		binder.noWait,
		binder.args,
	)
}

func Bind(queue string, options ...Option) Binder {
	binder := Binder{queue: queue}

	for _, option := range options {
		if option == nil {
			continue
		}

		option(&binder)
	}

	return binder
}

type Option func(*Binder)

func Title(title string) Option {
	return func(binder *Binder) { binder.title = title }
}

func Description(description string) Option {
	return func(binder *Binder) { binder.description = description }
}

func AutoAck(binder *Binder) { binder.autoAck = true }

func Exclusive(binder *Binder) { binder.exclusive = true }

func NoLocal(binder *Binder) { binder.noLocal = true }

func NoWait(binder *Binder) { binder.noWait = true }

func UseDedicatedChannel(binder *Binder) { binder.useNewChannel = true }
