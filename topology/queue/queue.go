package queue

import "github.com/streadway/amqp"

// type Channel interface {
// 	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
// 	QueueBind(name, exchange, routingKey string, noWait bool, args amqp.Table) error
// }

type Declarer struct {
	name        string
	description string

	exchange   string
	routingKey string
	shouldBind bool

	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool

	args amqp.Table

	deadLetterQueue *Declarer
}

func (d Declarer) Declare(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(d.name, d.durable, d.autoDelete, d.exclusive, d.noWait, d.args)
	if err != nil {
		return err
	}

	if d.shouldBind {
		err = ch.QueueBind(d.name, d.routingKey, d.exchange, d.noWait, nil)
		if err != nil {
			return err
		}
	}

	if dlq := d.deadLetterQueue; dlq != nil {
		err = dlq.Declare(ch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Declarer) addToTable(key string, value interface{}) {
	if d.args == nil {
		d.args = make(amqp.Table)
	}

	d.args[key] = value
}

func Declare(name string, options ...Option) Declarer {
	queue := Declarer{name: name}

	for _, option := range options {
		if option == nil {
			continue
		}

		option(&queue)
	}

	return queue
}

type Option func(*Declarer)

func Description(desc string) Option {
	return func(queue *Declarer) { queue.description = desc }
}

func BindTo(exchange, routingKey string) Option {
	return func(queue *Declarer) {
		queue.exchange = exchange
		queue.routingKey = routingKey
		queue.shouldBind = true
	}
}

func Durable(queue *Declarer) { queue.durable = true }

func AutoDelete(queue *Declarer) { queue.autoDelete = true }

func Exclusive(queue *Declarer) { queue.exclusive = true }

func NoWait(queue *Declarer) { queue.noWait = true }

func Arguments(args amqp.Table) Option {
	return func(queue *Declarer) {
		for key, value := range args {
			queue.addToTable(key, value)
		}
	}
}

func DeadLetter(exchange, routingKey string) Option {
	return func(queue *Declarer) {
		queue.addToTable("x-dead-letter-exchange", exchange)
		queue.addToTable("x-dead-letter-routing-key", routingKey)
	}
}

func DeadLetterWithQueue(exchange, routingKey string, dlq Declarer) Option {
	return func(queue *Declarer) {
		DeadLetter(exchange, routingKey)(queue)
		BindTo(exchange, routingKey)(&dlq)

		queue.deadLetterQueue = &dlq
	}
}
