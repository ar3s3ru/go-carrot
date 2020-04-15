package router_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/handler/router"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestRouter_Handle(t *testing.T) {
	t.Run("sending a Delivery on an empty router fails with router.ErrNoHandler", func(t *testing.T) {
		r := router.New()
		err := r.Handle(context.Background(), amqp.Delivery{ConsumerTag: "test-queue"})
		assert.True(t, errors.Is(err, router.ErrNoHandler))
	})

	t.Run("sending a Delivery with a binded Handler returns the handler error", func(t *testing.T) {
		r := router.New()
		r.Bind("test-queue", handler.Func(func(context.Context, amqp.Delivery) error {
			return nil
		}))

		err := r.Handle(context.Background(), amqp.Delivery{ConsumerTag: "test-queue"})
		assert.NoError(t, err)
	})

	t.Run("Use applies to the whole Router tree", func(t *testing.T) {
		calledTimes := 0

		r := router.New()
		r.Use(func(next handler.Handler) handler.Handler {
			return handler.Func(func(ctx context.Context, d amqp.Delivery) error {
				calledTimes++
				return next.Handle(ctx, d)
			})
		})

		r.Bind("test-queue", handler.Func(func(context.Context, amqp.Delivery) error {
			return nil
		}))

		r.Group(func(r router.Router) {
			r.Use(func(next handler.Handler) handler.Handler {
				return handler.Func(func(ctx context.Context, d amqp.Delivery) error {
					calledTimes += 10
					return next.Handle(ctx, d)
				})
			})

			r.Bind("test-queue-2", handler.Func(func(context.Context, amqp.Delivery) error {
				return nil
			}))
		})

		// First delivery: the calledTimes variable should only be updated by the
		// outer-most Use middleware.
		firstDelivery := amqp.Delivery{ConsumerTag: "test-queue"}
		err := r.Handle(context.Background(), firstDelivery)
		assert.NoError(t, err)
		assert.Equal(t, 1, calledTimes)

		calledTimes = 0

		// Second delivery: calledTimes will be updated by both Use middlewares.
		secondDelivery := amqp.Delivery{ConsumerTag: "test-queue-2"}
		err = r.Handle(context.Background(), secondDelivery)
		assert.NoError(t, err)
		assert.Equal(t, 11, calledTimes)
	})

	t.Run("With allows to specify middlewares for some Bind target", func(t *testing.T) {
		calledTimes := 0

		r := router.New()

		r.Bind("test-queue", handler.Func(func(context.Context, amqp.Delivery) error {
			calledTimes++
			return nil
		}))

		r.With(func(next handler.Handler) handler.Handler {
			return handler.Func(func(ctx context.Context, d amqp.Delivery) error {
				calledTimes += 10
				return next.Handle(ctx, d)
			})
		}).Bind("test-queue-2", handler.Func(func(context.Context, amqp.Delivery) error {
			return nil
		}))

		// First delivery: the calledTimes variable should only be updated by the
		// outer-most Use middleware.
		firstDelivery := amqp.Delivery{ConsumerTag: "test-queue"}
		err := r.Handle(context.Background(), firstDelivery)
		assert.NoError(t, err)
		assert.Equal(t, 1, calledTimes)

		calledTimes = 0

		// Second delivery: calledTimes will be updated by both Use middlewares.
		secondDelivery := amqp.Delivery{ConsumerTag: "test-queue-2"}
		err = r.Handle(context.Background(), secondDelivery)
		assert.NoError(t, err)
		assert.Equal(t, 10, calledTimes)
	})
}
