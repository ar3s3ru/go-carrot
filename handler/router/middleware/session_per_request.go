package middleware

import (
	"context"
	"database/sql"

	"github.com/ar3s3ru/go-carrot/handler"

	"github.com/streadway/amqp"
)

func SessionPerRequest(db *sql.DB) func(handler.Handler) handler.Handler {
	return func(next handler.Handler) handler.Handler {
		return handler.Func(func(ctx context.Context, delivery amqp.Delivery) error {
			panic("implement me")
		})
	}
}
