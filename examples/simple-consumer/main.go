package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/ar3s3ru/go-carrot"
	"github.com/ar3s3ru/go-carrot/binder"
	"github.com/ar3s3ru/go-carrot/binder/consumer"
	"github.com/ar3s3ru/go-carrot/handler"
	"github.com/ar3s3ru/go-carrot/handler/router"
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
			queue.Declare("consumer.message.received", queue.BindTo("messages", "message.published")),
			queue.Declare("consumer.message.deleted",
				queue.BindTo("messages", "message.published"),
				queue.BindTo("messages", "message.deleted"),
			),
		)),
		carrot.WithBinder(binder.All(
			consumer.Bind(
				"consumer.message.received",
				consumer.Title("Message received"),
				consumer.UseDedicatedChannel,
				// consumer.AutoAck,
			),
			consumer.Bind(
				"consumer.message.deleted",
				consumer.Title("Message deleted"),
			),
		)),
		carrot.WithHandler(router.New().Group(func(r router.Router) {
			r.Bind("consumer.message.received", LogMessages("(received) ", logger))
			r.Bind("consumer.message.deleted", LogMessages("(deleted) ", logger))
		})),
	).Run(), logger)

	<-time.After(config.App.Wait)
	logger.Printf("Stopping consumer after %s. See ya!\n", time.Since(start))
}

func LogMessages(prefix string, logger *log.Logger) handler.Handler {
	return handler.Func(func(ctx context.Context, delivery amqp.Delivery) error {
		logger.Println(
			prefix,
			"MessageId:", delivery.MessageId,
			"- Body:", string(delivery.Body),
		)

		return nil
	})
}

func mustNotFail(err error, logger *log.Logger) {
	if err != nil {
		logger.Fatalln("Unexpected error:", err)
	}
}
