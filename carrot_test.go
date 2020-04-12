package carrot_test

import (
	"context"
	"testing"

	"github.com/ar3s3ru/go-carrot"
	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/handler/router"
	"github.com/ar3s3ru/go-carrot/handler/router/middleware"
	"github.com/ar3s3ru/go-carrot/listener/consumer"
	"github.com/ar3s3ru/go-carrot/topology"
	"github.com/ar3s3ru/go-carrot/topology/exchange"
	"github.com/ar3s3ru/go-carrot/topology/exchange/kind"
	"github.com/ar3s3ru/go-carrot/topology/queue"

	"github.com/streadway/amqp"
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
