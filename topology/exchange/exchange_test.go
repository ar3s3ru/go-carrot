package exchange

import (
	"testing"

	"github.com/ar3s3ru/go-carrot/topology/exchange/kind"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestDeclare(t *testing.T) {
	testcases := map[string]struct {
		name    string
		options []Option
		output  Declarer
	}{
		"default has specified name and 'topic' kind": {
			name: "exchange",
			output: Declarer{
				name: "exchange",
				kind: kind.Topic,
			},
		},
		"turn up all the exchange options": {
			name: "exchange",
			options: []Option{
				Durable,
				AutoDelete,
				Exclusive,
				NoWait,
			},
			output: Declarer{
				name:       "exchange",
				kind:       kind.Topic,
				durable:    true,
				autoDelete: true,
				exclusive:  true,
				noWait:     true,
			},
		},
		"with arguments": {
			name: "exchange",
			options: []Option{
				Arguments(amqp.Table{
					"argument1": "value1",
					"argument2": "value2",
				}),
			},
			output: Declarer{
				name: "exchange",
				kind: kind.Topic,
				args: amqp.Table{
					"argument1": "value1",
					"argument2": "value2",
				},
			},
		},
		"binds to internal exchange and routingKey": {
			name: "exchange",
			options: []Option{
				BindTo("source", "routingKey"),
			},
			output: Declarer{
				name: "exchange",
				kind: kind.Topic,
				bindings: []binding{{
					exchange:   "source",
					routingKey: "routingKey",
				}},
			},
		},
		"multiple bindings": {
			name: "exchange",
			options: []Option{
				BindTo("source", "routingKey"),
				BindTo("source2", "routingKey"),
			},
			output: Declarer{
				name: "exchange",
				kind: kind.Topic,
				bindings: []binding{{
					exchange:   "source",
					routingKey: "routingKey",
				}, {
					exchange:   "source2",
					routingKey: "routingKey",
				}},
			},
		},
	}

	for name, tc := range testcases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) { assert.Equal(t, tc.output, Declare(tc.name, tc.options...)) })
	}
}
