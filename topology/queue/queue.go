package queue

import "github.com/streadway/amqp"

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

func (queue Declarer) Declare(ch *amqp.Channel) error {
	_, err := ch.QueueDeclare(queue.name, queue.durable, queue.autoDelete, queue.exclusive, queue.noWait, queue.args)
	if err != nil {
		return err
	}

	if queue.shouldBind {
		err = ch.QueueBind(queue.name, queue.routingKey, queue.exchange, queue.noWait, nil)
		if err != nil {
			return err
		}
	}

	if dlq := queue.deadLetterQueue; dlq != nil {
		err = dlq.Declare(ch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (queue *Declarer) addToTable(key string, value interface{}) {
	if queue.args == nil {
		queue.args = make(amqp.Table)
	}

	queue.args[key] = value
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
