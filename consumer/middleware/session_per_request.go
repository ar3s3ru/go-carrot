package middleware

import (
	"context"
	"database/sql"

	"github.com/ar3s3ru/go-carrot/consumer"

	"github.com/streadway/amqp"
)

func SessionPerRequest(db *sql.DB) func(consumer.Handler) consumer.Handler {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery amqp.Delivery) error {
			panic("implement me")
		})
	}
}
