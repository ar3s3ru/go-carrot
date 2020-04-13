// Package queue adds a topology.Declarer interface able to describe
// AMQP queues.
package queue

import (
	"github.com/ar3s3ru/go-carrot/topology"

	"github.com/streadway/amqp"
)

// Declarer is a topology component able to declare AMQP queues.
// Use Declare function to create a new instance of this component.
type Declarer struct {
	name        string
	description string

	bindings []binding

	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool

	args amqp.Table

	deadLetterQueue *Declarer
}

// Declare declares the topology of the queue using the supplied AMQP channel.
func (d Declarer) Declare(ch topology.Channel) error {
	_, err := ch.QueueDeclare(d.name, d.durable, d.autoDelete, d.exclusive, d.noWait, d.args)
	if err != nil {
		return err
	}

	if len(d.bindings) > 0 {
		if err = d.bindAll(ch); err != nil {
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

type binding struct {
	exchange   string
	routingKey string
}

func (d Declarer) bindAll(ch topology.Channel) error {
	for _, binding := range d.bindings {
		err := ch.QueueBind(d.name, binding.routingKey, binding.exchange, d.noWait, nil)
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

// Declare returns a new Declarer component able to declare the described AMQP queue.
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

// Option is an optional functionality that can be added to the Declarer
// that is being initialized by the Declare factory method.
type Option func(*Declarer)

// Description adds a description for the queue.
func Description(desc string) Option {
	return func(queue *Declarer) { queue.description = desc }
}

// BindTo describes a binding for the queue.
// Multiple calls of this option are supported.
func BindTo(exchange, routingKey string) Option {
	return func(queue *Declarer) {
		queue.bindings = append(queue.bindings, binding{
			exchange:   exchange,
			routingKey: routingKey,
		})
	}
}

// Durable will make the described queue survive AMQP broker restarts.
func Durable(queue *Declarer) { queue.durable = true }

// AutoDelete will automatically delete the described queue if there are no more
// consumers subscribed to the queue.
func AutoDelete(queue *Declarer) { queue.autoDelete = true }

// Exclusive will make the described queue be usable by only one AMQP connection
// and it will be deleted when such connection gets closed.
func Exclusive(queue *Declarer) { queue.exclusive = true }

// NoWait does not wait for the AMQP broker to confirm if queue declaration
// has succeeded, but the AMQP channel will instead assume the queue has been
// declared successfully.
func NoWait(queue *Declarer) { queue.noWait = true }

// Arguments specifies optional arguments to be supplied during queue declaration.
// Multiple calls of this option are supported.
func Arguments(args amqp.Table) Option {
	return func(queue *Declarer) {
		for key, value := range args {
			queue.addToTable(key, value)
		}
	}
}

// DeadLetter adds Dead Letter Exchange functionality to the described queue,
// by publishing rejected messages on the exchange and routingKey provided.
func DeadLetter(exchange, routingKey string) Option {
	return Arguments(amqp.Table{
		"x-dead-letter-exchange":    exchange,
		"x-dead-letter-routing-key": routingKey,
	})
}

// DeadLetterWithQueue declares a Dead Letter Exchange and a queue that
// will be binded to the specified exchange and routing key.
//
// Useful to persist failed messages in a specified queue and consuming
// messages from such queue from the application with a compensating action.
func DeadLetterWithQueue(exchange, routingKey string, dlq Declarer) Option {
	return func(queue *Declarer) {
		DeadLetter(exchange, routingKey)(queue)
		BindTo(exchange, routingKey)(&dlq)

		queue.deadLetterQueue = &dlq
	}
}
