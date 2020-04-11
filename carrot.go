package carrot

import (
	"github.com/ar3s3ru/go-carrot/consumer"
	"github.com/ar3s3ru/go-carrot/topology"

	"github.com/streadway/amqp"
)

func From(conn *amqp.Connection) Builder {
	return Builder{conn: conn}
}

type Builder struct {
	conn     *amqp.Connection
	declarer topology.Declarer
	router   consumer.Router
}

func (builder Builder) WithTopology(declarer topology.Declarer) Builder {
	builder.declarer = declarer
	return builder
}

func (builder Builder) WithConsumers(router consumer.Router) Builder {
	builder.router = router
	return builder
}

func (builder Builder) Start() {
}
