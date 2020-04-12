package listener

import (
	"fmt"

	"github.com/streadway/amqp"
)

func UseDedicatedChannel(listener Listener) Listener {
	if listener == nil {
		return nil
	}

	return Func(func(conn *amqp.Connection, _ *amqp.Channel) (<-chan amqp.Delivery, error) {
		ch, err := conn.Channel()
		if err != nil {
			return nil, fmt.Errorf("listener.UseDedicatedChannel: failed to open new channel, %w", err)
		}

		return listener.Listen(conn, ch)
	})
}
