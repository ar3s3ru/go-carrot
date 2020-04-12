package topology

// Declarer is a component able to declare a topology, given an AMQP channel
// to communicate with the AMQP broker.
//
// Declarer is a fallible operation, since it communicates with the external
// broker, so it may return an error.
type Declarer interface {
	Declare(Channel) error
}

// NOTE: not sure if this type should be exported...
type declarerFunc func(Channel) error

func (df declarerFunc) Declare(ch Channel) error { return df(ch) }
