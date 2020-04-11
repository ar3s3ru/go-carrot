package exchange

import (
	"github.com/ar3s3ru/go-carrot/topology/exchange/kind"

	"github.com/streadway/amqp"
)

type Declarer struct {
	name string
	kind kind.Kind

	source     string
	routingKey string
	shouldBind bool

	durable    bool
	autoDelete bool
	exclusive  bool
	noWait     bool

	args amqp.Table
}

func (exchange Declarer) Declare(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare(
		exchange.name,
		string(exchange.kind),
		exchange.durable, exchange.autoDelete,
		exchange.exclusive,
		exchange.noWait,
		exchange.args,
	)

	if err != nil {
		return err
	}

	if exchange.shouldBind {
		err = ch.ExchangeBind(exchange.name, exchange.routingKey, exchange.source, exchange.noWait, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (exchange *Declarer) addToTable(key string, value interface{}) {
	if exchange.args == nil {
		exchange.args = make(amqp.Table)
	}

	exchange.args[key] = value
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
		exchange.source = source
		exchange.routingKey = routingKey
		exchange.shouldBind = true
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
