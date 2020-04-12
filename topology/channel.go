package topology

import "github.com/streadway/amqp"

// Channel is the channel interface the topology Declarer uses
// to declare topologies.
type Channel interface {
	Tx() error
	TxCommit() error
	TxRollback() error

	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error

	ExchangeBind(destination, key, source string, noWait bool, args amqp.Table) error
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
}
