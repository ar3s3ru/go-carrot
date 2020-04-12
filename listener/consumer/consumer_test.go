package consumer

import (
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {
	testcases := map[string]struct {
		queue   string
		options []Option
		output  Listener
	}{
		"no options, just keeps the queue": {
			queue: "queue",
			output: Listener{
				queue: "queue",
			},
		},
		"nil options are ignored": {
			queue:   "queue",
			options: []Option{nil},
			output: Listener{
				queue: "queue",
			},
		},
		"with title and description": {
			queue: "queue",
			options: []Option{
				Title("my consumer"),
				Description("this is just a test consumer"),
			},
			output: Listener{
				queue:       "queue",
				title:       "my consumer",
				description: "this is just a test consumer",
			},
		},
		"turn up all the queue options": {
			queue: "queue",
			options: []Option{
				AutoAck,
				Exclusive,
				NoLocal,
				NoWait,
			},
			output: Listener{
				queue:     "queue",
				autoAck:   true,
				exclusive: true,
				noLocal:   true,
				noWait:    true,
			},
		},
		"with arguments": {
			queue: "queue",
			options: []Option{
				Arguments(amqp.Table{
					"argument1": "value1",
					"argument2": "value2",
				}),
			},
			output: Listener{
				queue: "queue",
				args: amqp.Table{
					"argument1": "value1",
					"argument2": "value2",
				},
			},
		},
	}

	for queue, tc := range testcases {
		queue, tc := queue, tc
		t.Run(queue, func(t *testing.T) { assert.Equal(t, tc.output, Listen(tc.queue, tc.options...)) })
	}
}
