package exchange

import (
	"github.com/ar3s3ru/go-carrot/topology/exchange/kind"

	"github.com/streadway/amqp"
)

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

func (d Declarer) Declare(ch *amqp.Channel) error {
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

func (d Declarer) bindAll(ch *amqp.Channel) error {
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

type Option func(*Declarer)

func BindTo(source, routingKey string) Option {
	return func(exchange *Declarer) {
		exchange.bindings = append(exchange.bindings, binding{
			exchange:   source,
			routingKey: routingKey,
		})
	}
}

func Kind(kind kind.Kind) Option {
	return func(exchange *Declarer) { exchange.kind = kind }
}

func Durable(exchange *Declarer) { exchange.durable = true }

func AutoDelete(exchange *Declarer) { exchange.autoDelete = true }

func Exclusive(exchange *Declarer) { exchange.exclusive = true }

func NoWait(exchange *Declarer) { exchange.noWait = true }

func Arguments(args amqp.Table) Option {
	return func(exchange *Declarer) {
		for key, value := range args {
			exchange.addToTable(key, value)
		}
	}
}
