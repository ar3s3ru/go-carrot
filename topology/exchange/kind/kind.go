// Package kind contains all supported AMQP exchange kinds.
package kind

// Kind represents an AMQP exchange kind.
type Kind string

// Supported AMQP exchange kinds.
//
// For more information, please visit https://www.rabbitmq.com/tutorials/amqp-concepts.html#exchanges.
const (
	Direct   Kind = "direct"
	Fanout   Kind = "fanout"
	Headers  Kind = "headers"
	Internal Kind = "internal"
	Topic    Kind = "topic"
)

func (kind Kind) String() string { return string(kind) }
