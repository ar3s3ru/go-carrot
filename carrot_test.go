package carrot_test

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/ar3s3ru/go-carrot"
	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/handler/router"
	"github.com/ar3s3ru/go-carrot/handler/router/middleware"
	"github.com/ar3s3ru/go-carrot/listener"
	"github.com/ar3s3ru/go-carrot/listener/consumer"
	"github.com/ar3s3ru/go-carrot/listener/mocks"
	"github.com/ar3s3ru/go-carrot/topology"
	"github.com/ar3s3ru/go-carrot/topology/exchange"
	"github.com/ar3s3ru/go-carrot/topology/exchange/kind"
	"github.com/ar3s3ru/go-carrot/topology/queue"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFrom(t *testing.T) {
	carrot.From(nil,
		carrot.WithTopology(topology.All(
			exchange.Declare("orders",
				exchange.Kind(kind.Topic),
				exchange.Durable,
			),
			exchange.Declare("orders-internal",
				exchange.Kind(kind.Topic),
				exchange.Durable,
			),
			queue.Declare(
				"my-service.order.invalidate",
				queue.BindTo("orders", "*.order.changed"),
				queue.Durable,
				queue.DeadLetterWithQueue(
					"orders", "my-service.order.invalidate.dead",
					queue.Declare(
						"my-service.order.invalidate.failed",
						queue.Durable,
						queue.NoWait,
					),
				),
			),
			queue.Declare(
				"my-service.order.finalized",
				queue.BindTo("orders", "*.order.finalized"),
				queue.Durable,
				queue.DeadLetterWithQueue(
					"orders", "my-service.order.finalized.dead",
					queue.Declare("my-service.order.finalized.failed"),
				),
			),
		)),
		carrot.WithListener(consumer.Listen(
			"my-service.order.invalidate",
			consumer.Title("Invalidate Order"),
		)),
		carrot.WithHandler(router.New().Group(func(r router.Router) {
			r.Use(middleware.SessionPerRequest(nil))

			r.Bind("my-service.order.invalidate", handler.Func(func(context.Context, amqp.Delivery) error {
				return nil
			}))
		})),
	).Run()
}

func TestWithGracefulShutdown(t *testing.T) {
	conn := new(mocks.Connection)
	conn.On("Channel").Return(nil, nil)
	conn.On("Close").Return(nil)

	ch := make(chan error)

	closer, err := carrot.From(conn,
		carrot.WithListener(
			listener.Func(func(listener.Connection, listener.Channel, handler.Handler) (listener.Closer, error) {
				closer := new(mocks.Closer)
				closer.On("Closed").Once().Return((<-chan error)(ch))
				closer.
					On("Close", mock.Anything).
					Run(func(mock.Arguments) {
						ch <- nil
						close(ch)
					}).
					Return(nil)

				return closer, nil
			}),
		),
		carrot.WithHandler(
			handler.Func(func(context.Context, amqp.Delivery) error { return nil }),
		),
		carrot.WithGracefulShutdown(&carrot.Shutdown{
			// Use SIGUSR1 to signal shutdown, because `go test` already listens
			// to SIGNINT and that would make the test fail.
			Signals: []os.Signal{syscall.SIGUSR1},
		}),
	).Run()

	assert.NoError(t, err)

	closeCh := closer.Closed()

	for {
		select {
		// Signal more than once to make sure the graceful shutdown goroutine
		// has been spawned and it's ready to listen to signals.
		case <-time.Tick(100 * time.Millisecond):
			assert.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGUSR1))

		case err = <-closeCh:
			assert.NoError(t, err)
			return

		case <-time.After(1 * time.Second):
			assert.Fail(t, "failed to gracefully close after 1 second")
			return
		}
	}
}
