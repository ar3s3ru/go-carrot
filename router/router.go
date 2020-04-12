package router

import (
	"context"

	"github.com/ar3s3ru/go-carrot"
	"github.com/streadway/amqp"
)

type Router interface {
	carrot.Handler

	Bind(string, carrot.Handler)

	Use(middlewares ...func(carrot.Handler) carrot.Handler)
	With(middlewares ...func(carrot.Handler) carrot.Handler) Router
	Group(func(Router)) Router
}

func New() *Mux {
	return new(Mux)
}

type Mux struct {
	middlewares []func(carrot.Handler) carrot.Handler
	consumers   map[string]carrot.Handler
}

func (r Mux) Handle(ctx context.Context, delivery amqp.Delivery) error {
	handler, ok := r.consumers[delivery.ConsumerTag]
	if !ok {
		return delivery.Nack(false, true)
	}

	err := handler.Handle(ctx, delivery)
	if err != nil {
		delivery.Nack(false, true)
	} else {
		delivery.Ack(true)
	}

	return err
}

func (r *Mux) Bind(queue string, handler carrot.Handler) {
	if handler == nil {
		return
	}

	if r.consumers == nil {
		r.consumers = make(map[string]carrot.Handler)
	}

	r.consumers[queue] = handler
}

func (r *Mux) Use(middlewares ...func(carrot.Handler) carrot.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r Mux) With(middlewares ...func(carrot.Handler) carrot.Handler) Router {
	newRouter := Mux{consumers: r.consumers}
	newRouter.middlewares = append(r.middlewares, middlewares...)

	return &newRouter
}

func (r *Mux) Group(fn func(Router)) Router {
	newRouter := r.With()

	if fn != nil {
		fn(newRouter)
	}

	return newRouter
}

func (r Mux) applyMiddlewares(handler carrot.Handler) carrot.Handler {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		middleware := r.middlewares[i]
		handler = middleware(handler)
	}

	return handler
}
