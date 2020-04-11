package topology

import (
	"fmt"

	"github.com/streadway/amqp"
)

type Declarer interface {
	Declare(*amqp.Channel) error
}

type declarerFunc func(*amqp.Channel) error

func (df declarerFunc) Declare(ch *amqp.Channel) error { return df(ch) }

func All(first Declarer, others ...Declarer) Declarer {
	return declarerFunc(func(ch *amqp.Channel) error {
		others = append(others, first)
		for _, declarer := range others {
			if declarer == nil {
				continue
			}

			if err := declarer.Declare(ch); err != nil {
				return fmt.Errorf("topology.All: failed to declare topology %w", err)
			}
		}

		return nil
	})
}
