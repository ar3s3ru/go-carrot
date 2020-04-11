package carrot

import (
	"github.com/ar3s3ru/go-carrot/consumer"
	"github.com/ar3s3ru/go-carrot/consumer/middleware"
	"github.com/ar3s3ru/go-carrot/topology/queue"
)

func ExampleFrom() {
	From(nil).
		WithTopology(queue.Declare(
			"my-service.order.invalidate",
			queue.Durable,
			queue.DeadLetterWithQueue(
				"orders", "my-service.order.invalidate.dead",
				queue.Declare(
					"my-service.order.invalidate.failed",
					queue.Durable,
					queue.NoWait,
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
