package kind

type Kind string

const (
	Direct   Kind = "direct"
	Fanout   Kind = "fanout"
	Headers  Kind = "headers"
	Internal Kind = "internal"
	Topic    Kind = "topic"
)

func (kind Kind) String() string { return string(kind) }
