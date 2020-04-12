package listener

import (
	"fmt"
	"sync"

	"github.com/streadway/amqp"
)

// Sink allows for listening from multiple Listeners, by maintaining
// an amqp.Delivery sink to which all the messages are sent.
//
// Returns nil if no listeners are supplied.
func Sink(listeners ...Listener) Listener {
	if len(listeners) == 0 {
		return nil
	}

	sinker := new(sinker)
	sinker.listeners = listeners

	return sinker
}

type sinker struct {
	sync.WaitGroup

	listeners  []Listener
	deliveries []<-chan amqp.Delivery
}

func (sinker *sinker) Listen(conn *amqp.Connection, ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	if err := sinker.deliveriesSinks(conn, ch); err != nil {
		return nil, err
	}

	return sinker.spawnSinkersFromDeliveries(), nil
}

func (sinker *sinker) deliveriesSinks(conn *amqp.Connection, ch *amqp.Channel) error {
	deliveries := make([]<-chan amqp.Delivery, 0, len(sinker.listeners))

	for _, listener := range sinker.listeners {
		delivery, err := listener.Listen(conn, ch)
		if err != nil {
			return fmt.Errorf("listener.Sink: failed to listen, %w", err)
		}

		deliveries = append(deliveries, delivery)
	}

	sinker.deliveries = deliveries

	return nil
}

func (sinker *sinker) spawnSinkersFromDeliveries() <-chan amqp.Delivery {
	sinker.Add(len(sinker.deliveries))
	sink := make(chan amqp.Delivery, len(sinker.deliveries))

	for _, delivery := range sinker.deliveries {
		go sinker.sink(sink, delivery)
	}

	go func() {
		sinker.Wait()
		close(sink)
	}()

	return sink
}

func (sinker *sinker) sink(sink chan<- amqp.Delivery, source <-chan amqp.Delivery) {
	defer sinker.Done()
	for delivery := range source {
		sink <- delivery
	}
}
