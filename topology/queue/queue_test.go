package queue

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestDeclare(t *testing.T) {
	testcases := map[string]struct {
		name    string
		options []Option
		output  Declarer
	}{
		"no options, just keeps the name": {
			name: "queue",
			output: Declarer{
				name: "queue",
			},
		},
		"nil options are ignored": {
			name:    "queue",
			options: []Option{nil},
			output: Declarer{
				name: "queue",
			},
		},
		"with description": {
			name: "queue",
			options: []Option{
				Description("this is just a queue"),
			},
			output: Declarer{
				name:        "queue",
				description: "this is just a queue",
			},
		},
		"binds to exchange and routingKey": {
			name: "queue",
			options: []Option{
				BindTo("exchange", "routingKey"),
			},
			output: Declarer{
				name: "queue",
				bindings: []binding{{
					exchange:   "exchange",
					routingKey: "routingKey",
				}},
			},
		},
		"multiple bindings": {
			name: "queue",
			options: []Option{
				BindTo("exchange", "routingKey"),
				BindTo("exchange2", "routingKey"),
			},
			output: Declarer{
				name: "queue",
				bindings: []binding{{
					exchange:   "exchange",
					routingKey: "routingKey",
				}, {
					exchange:   "exchange2",
					routingKey: "routingKey",
				}},
			},
		},
		"turn up all the queue options": {
			name: "queue",
			options: []Option{
				Durable,
				AutoDelete,
				Exclusive,
				NoWait,
			},
			output: Declarer{
				name:       "queue",
				durable:    true,
				autoDelete: true,
				exclusive:  true,
				noWait:     true,
			},
		},
		"with arguments": {
			name: "queue",
			options: []Option{
				Arguments(amqp.Table{
					"argument1": "value1",
					"argument2": "value2",
				}),
			},
			output: Declarer{
				name: "queue",
				args: amqp.Table{
					"argument1": "value1",
					"argument2": "value2",
				},
			},
		},
		"enable dead letter": {
			name: "queue",
			options: []Option{
				DeadLetter("exchange", "routingKey"),
			},
			output: Declarer{
				name: "queue",
				args: amqp.Table{
					"x-dead-letter-exchange":    "exchange",
					"x-dead-letter-routing-key": "routingKey",
				},
			},
		},
		"enable dead letter with queue": {
			name: "queue",
			options: []Option{
				DeadLetterWithQueue(
					"exchange", "routingKey",
					Declare("dead-letter-queue"),
				),
			},
			output: Declarer{
				name: "queue",
				args: amqp.Table{
					"x-dead-letter-exchange":    "exchange",
					"x-dead-letter-routing-key": "routingKey",
				},
				deadLetterQueue: &Declarer{
					name: "dead-letter-queue",
					bindings: []binding{{
						exchange:   "exchange",
						routingKey: "routingKey",
					}},
				},
			},
		},
	}

	for name, tc := range testcases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) { assert.Equal(t, tc.output, Declare(tc.name, tc.options...)) })
	}
}
