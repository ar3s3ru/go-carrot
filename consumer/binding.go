package consumer

import "github.com/streadway/amqp"

type Binding struct {
	Title       string
	Description string
	Queue       string

	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Arguments amqp.Table
}

func (binding Binding) Bind(ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		binding.Queue,
		"",
		binding.AutoAck,
		binding.Exclusive,
		binding.NoLocal,
		binding.NoWait,
		binding.Arguments,
	)
}
