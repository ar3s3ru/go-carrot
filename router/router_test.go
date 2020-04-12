package router_test

import (
	"context"
	"testing"

	"github.com/ar3s3ru/go-carrot"
	"github.com/ar3s3ru/go-carrot/router"
	"github.com/ar3s3ru/go-carrot/router/middleware"

	"github.com/streadway/amqp"
)

func TestNew(t *testing.T) {
	router := router.New().Group(func(r router.Router) {
		r.Use(middleware.SessionPerRequest(nil))

		r.Bind("my-service.order.invalidate", carrot.HandlerFunc(func(context.Context, amqp.Delivery) error {
			return nil
		}))
	})

	t.Logf("router: %#v\n", router)
}
