package listener

import (
	"fmt"

	"github.com/ar3s3ru/go-carrot/handler"
)

// UseDedicatedChannel creates a new, dedicated amqp.Channel to be used
// exclusively by the wrapped Listener.
func UseDedicatedChannel(listener Listener) Listener {
	if listener == nil {
		return nil
	}

	return Func(func(conn Connection, _ Channel, h handler.Handler) (Closer, error) {
		ch, err := conn.Channel()
		if err != nil {
			return nil, fmt.Errorf("listener.UseDedicatedChannel: failed to open new channel, %w", err)
		}

		return listener.Listen(conn, ch, h)
	})
}
