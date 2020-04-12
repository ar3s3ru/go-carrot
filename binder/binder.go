package binder

import (
	"fmt"
	"sync"

	"github.com/streadway/amqp"
)

type Binder interface {
	Bind(*amqp.Connection, *amqp.Channel) (<-chan amqp.Delivery, error)
}

type Func func(*amqp.Connection, *amqp.Channel) (<-chan amqp.Delivery, error)

func (fn Func) Bind(conn *amqp.Connection, ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	return fn(conn, ch)
}

func All(binders ...Binder) Binder {
	if len(binders) == 0 {
		return nil
	}

	return Func(func(conn *amqp.Connection, ch *amqp.Channel) (<-chan amqp.Delivery, error) {
		deliveries := make([]<-chan amqp.Delivery, 0, len(binders))

		for _, binder := range binders {
			delivery, err := binder.Bind(conn, ch)
			if err != nil {
				return nil, fmt.Errorf("binder.All: failed to bind, %w", err)
			}

			deliveries = append(deliveries, delivery)
		}

		wg := new(sync.WaitGroup)
		wg.Add(len(deliveries))

		sink := make(chan amqp.Delivery)
		for _, delivery := range deliveries {
			go sinkFromSource(wg, sink, delivery)
		}

		go func() {
			wg.Wait()
			close(sink)
		}()

		return sink, nil
	})
}

func sinkFromSource(wg *sync.WaitGroup, sink chan<- amqp.Delivery, source <-chan amqp.Delivery) {
	for delivery := range source {
		sink <- delivery
	}

	wg.Done()
}
