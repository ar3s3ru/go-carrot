package consumer_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/listener/consumer"
	"github.com/ar3s3ru/go-carrot/listener/mocks"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListener_Server(t *testing.T) {
	t.Run("it closes the channel successfully", func(t *testing.T) {
		listener := consumer.Listen(
			"test-queue",
			consumer.OnSuccess(func(amqp.Delivery) {
				assert.Fail(t, "called success callback")
			}),
			consumer.OnError(func(_ amqp.Delivery, err error) {
				assert.Fail(t, "called error callback", err)
			}),
		)

		sink := make(chan amqp.Delivery)
		var sinkCloser sync.Once

		ch := new(mocks.Channel)
		ch.
			On("Consume", "test-queue", "test-queue", false, false, false, false, amqp.Table(nil)).
			Return((<-chan amqp.Delivery)(sink), nil)
		ch.
			On("Close").
			Run(func(args mock.Arguments) { sinkCloser.Do(func() { close(sink) }) }).
			Return(nil)

		closer, err := listener.Listen(nil, ch, handler.Func(func(context.Context, amqp.Delivery) error {
			return errors.New("should not be called")
		}))

		assert.NoError(t, err)
		assert.NotNil(t, closer)

		go func() {
			assert.NoError(t, closer.Close(context.Background()))
		}()

		select {
		case err := <-closer.Closed():
			assert.NoError(t, err)
		case <-time.After(1 * time.Second):
			assert.Fail(t, "did not finish after 1 second")
		}

		// Closing a second time will return an error
		assert.Equal(t, consumer.ErrAlreadyClosed, closer.Close(context.Background()))
	})

	t.Run("it calls the success callback when the message is being handled correctly", func(t *testing.T) {
		// Use this channel to mark completeness by sending a value from the
		// server callbacks.
		done := make(chan bool)
		defer close(done)

		received := false
		delivery := amqp.Delivery{
			ConsumerTag: "test-queue",
			DeliveryTag: 1,
		}

		listener := consumer.Listen(
			"test-queue",
			consumer.OnSuccess(func(success amqp.Delivery) {
				received = true
				assert.Equal(t, delivery, success)
				done <- true
			}),
			consumer.OnError(func(_ amqp.Delivery, err error) {
				assert.Fail(t, "called error callback", err)
				done <- true
			}),
		)

		sink := make(chan amqp.Delivery)
		defer close(sink)

		ch := new(mocks.Channel)
		ch.
			On("Consume", "test-queue", "test-queue", false, false, false, false, amqp.Table(nil)).
			Return((<-chan amqp.Delivery)(sink), nil)

		closer, err := listener.Listen(nil, ch, handler.Func(func(context.Context, amqp.Delivery) error {
			return nil
		}))

		assert.NoError(t, err)
		assert.NotNil(t, closer)

		sink <- delivery

		select {
		case <-done:
			assert.True(t, received)
		case <-time.After(1 * time.Second):
			assert.Fail(t, "did not finish after 1 second")
		}
	})

	t.Run("it calls the error callback when the message handling failed", func(t *testing.T) {
		// Use this channel to mark completeness by sending a value from the
		// server callbacks.
		done := make(chan bool)
		defer close(done)

		received := false
		delivery := amqp.Delivery{
			ConsumerTag: "test-queue",
			DeliveryTag: 1,
		}
		failure := errors.New("failed message")

		listener := consumer.Listen(
			"test-queue",
			consumer.OnSuccess(func(amqp.Delivery) {
				assert.Fail(t, "called success callback")
				done <- true
			}),
			consumer.OnError(func(failed amqp.Delivery, err error) {
				received = true
				assert.Equal(t, delivery, failed)
				assert.True(t, errors.Is(err, failure))
				done <- true
			}),
		)

		sink := make(chan amqp.Delivery)
		defer close(sink)

		ch := new(mocks.Channel)
		ch.
			On("Consume", "test-queue", "test-queue", false, false, false, false, amqp.Table(nil)).
			Return((<-chan amqp.Delivery)(sink), nil)

		closer, err := listener.Listen(nil, ch, handler.Func(func(context.Context, amqp.Delivery) error {
			// Fail message handling
			return failure
		}))

		assert.NoError(t, err)
		assert.NotNil(t, closer)

		sink <- delivery

		select {
		case <-done:
			assert.True(t, received)
		case <-time.After(1 * time.Second):
			assert.Fail(t, "did not finish after 1 second")
		}
	})
}
