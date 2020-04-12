package middleware

import (
	"context"
	"time"

	"github.com/ar3s3ru/go-carrot/handler"

	"github.com/streadway/amqp"
)

func Timeout(deadline time.Duration) func(handler.Handler) handler.Handler {
	return func(next handler.Handler) handler.Handler {
		return handler.Func(func(ctx context.Context, delivery amqp.Delivery) error {
			ctx, cancel := context.WithTimeout(ctx, deadline)
			defer cancel()

			return next.Handle(ctx, delivery)

			// ch := make(chan error, 1)
			// go func() {
			// 	defer close(ch)
			// 	ch <- next.Handle(ctx, delivery)
			// }()

			// select {
			// case <-ctx.Done():
			// 	return ctx.Err()
			// case err := <-ch:
			// 	return err
			// }
		})
	}
}
