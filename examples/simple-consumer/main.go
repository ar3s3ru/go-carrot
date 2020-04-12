package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/ar3s3ru/go-carrot"
	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/handler/router"
	"github.com/ar3s3ru/go-carrot/handler/router/middleware"
	"github.com/ar3s3ru/go-carrot/listener"
	"github.com/ar3s3ru/go-carrot/listener/consumer"
	"github.com/ar3s3ru/go-carrot/topology"
	"github.com/ar3s3ru/go-carrot/topology/exchange"
	"github.com/ar3s3ru/go-carrot/topology/queue"

	"github.com/kelseyhightower/envconfig"
	"github.com/streadway/amqp"
)

type Config struct {
	AMQP struct {
		Addr string `default:"amqp://guest:guest@rabbit:5672"`
	}

	App struct {
		Wait time.Duration `default:"1m"`
	}
}

func main() {
	logger := log.New(os.Stderr, "[go-carrot/simple-consumer] ", log.LstdFlags)

	var config Config
	mustNotFail(envconfig.Process("", &config), logger)

	conn, err := amqp.Dial(config.AMQP.Addr)
	mustNotFail(err, logger)

	start := time.Now()
	logger.Println("Starting consumers...")

	mustNotFail(carrot.From(conn,
		carrot.WithTopology(topology.All(
			exchange.Declare("messages"),
			queue.Declare("consumer.message.received",
				queue.BindTo("messages", "message.published"),
			),
			queue.Declare("consumer.message.deleted",
				queue.BindTo("messages", "message.published"),
				queue.BindTo("messages", "message.deleted"),
			),
		)),
		carrot.WithListener(listener.Sink(
			consumer.Listen(
				"consumer.message.received",
				consumer.Title("Message received"),
				consumer.UseDedicatedChannel,
			),
			consumer.Listen(
				"consumer.message.deleted",
				consumer.Title("Message deleted"),
			),
		)),
		carrot.WithHandler(router.New().Group(func(r router.Router) {
			r.Use(LogMessages(logger))
			r.Use(middleware.Timeout(50 * time.Millisecond))
			r.Use(SimulateWork(100*time.Millisecond, logger))

			r.Bind("consumer.message.received", handler.Func(Acknowledger))
			r.Bind("consumer.message.deleted", handler.Func(Acknowledger))
		})),
	).Run(), logger)

	<-time.After(config.App.Wait)
	logger.Printf("Stopping consumer after %s. See ya!\n", time.Since(start))
}

func SimulateWork(max time.Duration, logger *log.Logger) func(handler.Handler) handler.Handler {
	return func(next handler.Handler) handler.Handler {
		return handler.Func(func(ctx context.Context, delivery amqp.Delivery) error {
			waitingTime := time.Duration(rand.Int63()) % max
			logger.Printf("MessageId: %s - Waiting for '%s'\n", delivery.MessageId, waitingTime)

			<-time.After(waitingTime)

			return next.Handle(ctx, delivery)
		})
	}
}

func LogMessages(logger *log.Logger) func(handler.Handler) handler.Handler {
	return func(next handler.Handler) handler.Handler {
		return handler.Func(func(ctx context.Context, delivery amqp.Delivery) error {
			err := next.Handle(ctx, delivery)

			if err != nil {
				logger.Printf("(%s @ %s) Failed: %s\n",
					delivery.MessageId, delivery.ConsumerTag,
					err,
				)
			} else {
				logger.Printf("(%s @ %s) %s\n",
					delivery.MessageId, delivery.ConsumerTag,
					delivery.Body,
				)
			}

			return err
		})
	}
}

func Acknowledger(ctx context.Context, delivery amqp.Delivery) error {
	return nil
}

func mustNotFail(err error, logger *log.Logger) {
	if err != nil {
		logger.Fatalln("Unexpected error:", err)
	}
}
