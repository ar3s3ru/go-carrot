// Package exchange adds a topology.Declarer interface able to describe
// AMQP exchanges.
package exchange

import (
	"github.com/ar3s3ru/go-carrot/topology"
	"github.com/ar3s3ru/go-carrot/topology/exchange/kind"

	"github.com/streadway/amqp"
)

// Declarer is a topology component able to declare AMQP exchanges.
// Use Declare function to create a new instance of this component.
type Declarer struct {
	name string
	kind kind.Kind

	bindings []binding

	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool

	args amqp.Table
}

// Declare declares the topology of the exchange using the supplied AMQP channel.
func (d Declarer) Declare(ch topology.Channel) error {
	err := ch.ExchangeDeclare(d.name, string(d.kind), d.durable, d.autoDelete, d.exclusive, d.noWait, d.args)
	if err != nil {
		return err
	}

	if len(d.bindings) > 0 {
		if err := d.bindAll(ch); err != nil {
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
		err := ch.ExchangeBind(d.name, binding.routingKey, binding.exchange, d.noWait, nil)
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

// Declare returns a new Declarer component able to declare the described AMQP exchange.
func Declare(name string, options ...Option) Declarer {
	exchange := Declarer{name: name, kind: kind.Topic}

	for _, option := range options {
		if option == nil {
			continue
		}

		option(&exchange)
	}

	return exchange
}

// Option is an optional functionality that can be added to the Declarer
// that is being initialized by the Declare factory method.
type Option func(*Declarer)

// Kind specifies the exchange kind.
//
// For more information about exchange kinds,
// please visit https://www.rabbitmq.com/tutorials/amqp-concepts.html#exchanges.
func Kind(kind kind.Kind) Option {
	return func(exchange *Declarer) { exchange.kind = kind }
}

// BindTo binds the described exchange to another source exchange, redirecting
// messages published from the specified exchanged on the exchange being described.
//
// Multiple calls of this option are supported.
func BindTo(source, routingKey string) Option {
	return func(exchange *Declarer) {
		exchange.bindings = append(exchange.bindings, binding{
			exchange:   source,
			routingKey: routingKey,
		})
	}
}

// Durable will make the described exchange survive AMQP broker restarts.
func Durable(exchange *Declarer) { exchange.durable = true }

// AutoDelete will automatically delete the described exchange if there are no more
// queues subscribed to the exchange.
func AutoDelete(exchange *Declarer) { exchange.autoDelete = true }

// Exclusive will make the described exchange be usable by only one AMQP connection
// and it will be deleted when such connection gets closed.
func Exclusive(exchange *Declarer) { exchange.exclusive = true }

// NoWait does not wait for the AMQP broker to confirm if exchange declaration
// has succeeded, but the AMQP channel will instead assume the exchange has been
// declared successfully.
func NoWait(exchange *Declarer) { exchange.noWait = true }

// Arguments specifies optional arguments to be supplied during exchange declaration.
// Multiple calls of this option are supported.
func Arguments(args amqp.Table) Option {
	return func(exchange *Declarer) {
		for key, value := range args {
			exchange.addToTable(key, value)
		}
	}
}
