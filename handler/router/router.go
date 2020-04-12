package router

import (
	"context"

	"github.com/ar3s3ru/go-carrot/handler"

	"github.com/streadway/amqp"
)

type Router interface {
	handler.Handler

	Bind(string, handler.Handler)

	Use(middlewares ...func(handler.Handler) handler.Handler)
	With(middlewares ...func(handler.Handler) handler.Handler) Router
	Group(func(Router)) Router
}

func New() *Mux {
	return new(Mux)
}

type Mux struct {
	middlewares []func(handler.Handler) handler.Handler
	consumers   map[string]handler.Handler
}

func (r Mux) Handle(ctx context.Context, delivery amqp.Delivery) error {
	handler, ok := r.consumers[delivery.ConsumerTag]
	if !ok {
		return delivery.Nack(false, true)
	}

	err := r.applyMiddlewaresTo(handler).Handle(ctx, delivery)
	if err != nil {
		delivery.Nack(false, true)
		return err
	}

	return delivery.Ack(true)
}

func (r *Mux) Bind(queue string, h handler.Handler) {
	if h == nil {
		return
	}

	if r.consumers == nil {
		r.consumers = make(map[string]handler.Handler)
	}

	r.consumers[queue] = h
}

func (r *Mux) Use(middlewares ...func(handler.Handler) handler.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r Mux) With(middlewares ...func(handler.Handler) handler.Handler) Router {
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

func (r Mux) applyMiddlewaresTo(handler handler.Handler) handler.Handler {
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		middleware := r.middlewares[i]
		handler = middleware(handler)
	}

	return handler
}
