package listener

import (
	"github.com/streadway/amqp"
)

type Listener interface {
	Listen(*amqp.Connection, *amqp.Channel) (<-chan amqp.Delivery, error)
}

type Func func(*amqp.Connection, *amqp.Channel) (<-chan amqp.Delivery, error)

func (fn Func) Listen(conn *amqp.Connection, ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	return fn(conn, ch)
}
