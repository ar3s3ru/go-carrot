package carrot_test

import (
	"testing"

	"github.com/ar3s3ru/go-carrot"
	"github.com/ar3s3ru/go-carrot/consumer"
	"github.com/ar3s3ru/go-carrot/consumer/middleware"
	"github.com/ar3s3ru/go-carrot/topology"
	"github.com/ar3s3ru/go-carrot/topology/exchange"
	"github.com/ar3s3ru/go-carrot/topology/exchange/kind"
	"github.com/ar3s3ru/go-carrot/topology/queue"
)

func TestFrom(t *testing.T) {
	carrot.From(nil).
		WithTopology(topology.All(
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
		)).
		WithConsumers(consumer.NewRouter().Group(func(r consumer.Router) {
			r.Use(middleware.SessionPerRequest(nil))

			r.Register(nil, consumer.Binding{
				Title: "Invalidate Orders",
				Queue: "my-service.order.invalidate",
			})
		})).
		Start()
}
