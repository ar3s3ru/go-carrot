package middleware

import (
	"context"
	"database/sql"

	"github.com/ar3s3ru/go-carrot"

	"github.com/streadway/amqp"
)

func SessionPerRequest(db *sql.DB) func(carrot.Handler) carrot.Handler {
	return func(next carrot.Handler) carrot.Handler {
		return carrot.HandlerFunc(func(ctx context.Context, delivery amqp.Delivery) error {
			panic("implement me")
		})
	}
}
