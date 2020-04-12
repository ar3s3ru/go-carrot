package listener

import (
	"fmt"
	"sync"

	"github.com/streadway/amqp"
)

func Sink(listeners ...Listener) Listener {
	if len(listeners) == 0 {
		return nil
	}

	return Func(func(conn *amqp.Connection, ch *amqp.Channel) (<-chan amqp.Delivery, error) {
		deliveries := make([]<-chan amqp.Delivery, 0, len(listeners))

		for _, listener := range listeners {
			delivery, err := listener.Listen(conn, ch)
			if err != nil {
				return nil, fmt.Errorf("listener.Sink: failed to listen, %w", err)
			}

			deliveries = append(deliveries, delivery)
		}

		wg := new(sync.WaitGroup)
		wg.Add(len(deliveries))

		sink := make(chan amqp.Delivery, 1)
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
		fmt.Println("sinking delivery from", delivery.ConsumerTag)
		sink <- delivery
	}

	wg.Done()
}
