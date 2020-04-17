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

// Config is the configuration used by this executable.
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

	closer, err := carrot.Run(conn,
		// Declare the application topology, making sure it matches with the
		// one present on the AMQP broker. If not, carrot will fail execution.
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
		// Listen for incoming messages on multiple consumers by using listener.Sink.
		carrot.WithListener(listener.Sink(
			consumer.Listen("consumer.message.received"),
			consumer.Listen("consumer.message.deleted"),
		)),
		// Use a Router as message handler function to route messages from both
		// queues to the appropriate handler function, and add some
		// middlewares on top of them.
		carrot.WithHandler(router.New().Group(func(r router.Router) {
			r.Use(LogMessages(logger))
			r.Use(middleware.Timeout(50 * time.Millisecond))
			r.Use(SimulateWork(100*time.Millisecond, logger))

			r.Bind("consumer.message.received", handler.Func(Acknowledger))
			r.Bind("consumer.message.deleted", handler.Func(Acknowledger))
		})),
		// Enables graceful shutdown when an interrupt signal is received.
		carrot.WithGracefulShutdown(nil),
	)

	mustNotFail(err, logger)

	// Close on the background.
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := closer.Close(ctx); err != nil {
			logger.Fatalf("Consumers closed with error: %s", err)
		}

		logger.Println("Consumers closed successfully")
	}()

	select {
	case err := <-closer.Closed():
		logger.Fatalf("Consumers closed with error: %s", err)
	case <-time.After(config.App.Wait):
		logger.Printf("Stopping consumer after %s. See ya!", time.Since(start))
	}
}

// SimulateWork simulates some delay between receiving the message and handing
// the message out to the handler function.
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

// LogMessages logs the incoming message, after being handled by the underlying
// handler function.
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

// Acknowledger force message acknowledgement by returning no error.
func Acknowledger(context.Context, amqp.Delivery) error { return nil }

func mustNotFail(err error, logger *log.Logger) {
	if err != nil {
		logger.Fatalln("Unexpected error:", err)
	}
}
