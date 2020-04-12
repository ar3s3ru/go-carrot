package router

import (
	"context"

	"github.com/ar3s3ru/go-carrot/handler"

	"github.com/streadway/amqp"
)

// Router is a message handler extension that supports middlewares
// and multiple consumer bindings.
type Router interface {
	handler.Handler

	Bind(string, handler.Handler)

	Use(middlewares ...func(handler.Handler) handler.Handler)
	With(middlewares ...func(handler.Handler) handler.Handler) Router
	Group(func(Router)) Router
}

// New returns a new Router instance.
func New() *Mux {
	return new(Mux)
}

// Mux is a multiplexer that implements the Router interface,
// to support multiple message handler functions for specific queues.
type Mux struct {
	middlewares []func(handler.Handler) handler.Handler
	consumers   map[string]handler.Handler
}

// Handle delegates message handling to the specific queue identified by
// the amqp.Delivery.ConsumerTag value.
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

// Bind binds a message handler function to the specified queue, if the
// handler is not nil.
func (r *Mux) Bind(queue string, h handler.Handler) {
	if h == nil {
		return
	}

	if r.consumers == nil {
		r.consumers = make(map[string]handler.Handler)
	}

	r.consumers[queue] = h
}

// Use appends middlewares to the Mux middleware stack.
func (r *Mux) Use(middlewares ...func(handler.Handler) handler.Handler) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// With adds inline middlewares for a message handler, returning the new
// Router instance with the new middlewares appended to the Mux middleware stack.
func (r Mux) With(middlewares ...func(handler.Handler) handler.Handler) Router {
	newRouter := Mux{consumers: r.consumers}
	newRouter.middlewares = append(r.middlewares, middlewares...)

	return &newRouter
}

// Group creates a new inline Router and fresh middleware stack, useful to group
// multiple handler bindings with same middlewares to be applied.
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
