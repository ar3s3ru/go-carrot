package router

import (
	"context"
	"errors"

	"github.com/ar3s3ru/go-carrot/handler"

	"github.com/streadway/amqp"
)

// ErrNoHandler is returned by Router.Handle method when no handler
// has been found for the incoming message.
var ErrNoHandler = errors.New("router: no handler for the incoming message has been found")

// Router is a message handler extension that supports middlewares
// and multiple consumer bindings.
type Router interface {
	Binder
	handler.Handler

	Group(func(Router)) Router
	Use(middlewares ...func(handler.Handler) handler.Handler)
	With(middlewares ...func(handler.Handler) handler.Handler) Binder
}

// Binder allows to bind a queue to a message handler.
type Binder interface {
	Bind(string, handler.Handler)
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
		return ErrNoHandler
	}

	return applyTo(handler, r.middlewares...).Handle(ctx, delivery)
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

// Group creates a new inline Router and fresh middleware stack, useful to group
// multiple handler bindings with same middlewares to be applied.
func (r *Mux) Group(fn func(Router)) Router {
	newRouter := New()

	if fn != nil {
		fn(newRouter)
	}

	for consumer := range newRouter.consumers {
		r.Bind(consumer, newRouter)
	}

	return r
}

// With adds inline middlewares for a message handler, returning a Binder interface
// that can be used to bind the message handler for a specific consumer.
func (r *Mux) With(middlewares ...func(handler.Handler) handler.Handler) Binder {
	return delegatedBinder{
		router:      r,
		middlewares: middlewares,
	}
}

type delegatedBinder struct {
	router      *Mux
	middlewares []func(handler.Handler) handler.Handler
}

func (b delegatedBinder) Bind(name string, h handler.Handler) {
	b.router.Bind(name, applyTo(h, b.middlewares...))
}

func applyTo(handler handler.Handler, middlewares ...func(handler.Handler) handler.Handler) handler.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		middleware := middlewares[i]
		handler = middleware(handler)
	}

	return handler
}
