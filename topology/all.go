package topology

import (
	"fmt"

	"github.com/streadway/amqp"
)

// All declares the whole topology provided in a transaction.
//
// An error is returned if committing, rolling-back or just declaring the topology
// failed with an error.
//
// A nil Declarer is returned instead if no Declarers are supplied as arguments.
func All(declarers ...Declarer) Declarer {
	if len(declarers) == 0 {
		return nil
	}

	return declarerFunc(func(ch *amqp.Channel) error {
		var err error

		if err = ch.Tx(); err != nil {
			return fmt.Errorf("topology.All: failed to open transaction on channel, %w", err)
		}

		// Rollbacks the transaction in case the topology declaration has failed.
		defer func() {
			if err == nil {
				return
			}

			if rollbackErr := ch.TxRollback(); rollbackErr != nil {
				err = fmt.Errorf("topology.All: failed to rollback transaction, %w (caused by %s)",
					rollbackErr,
					err,
				)
			}
		}()

		for _, declarer := range declarers {
			if declarer == nil {
				continue
			}

			if err = declarer.Declare(ch); err != nil {
				err = fmt.Errorf("topology.All: failed to declare topology, %w", err)
				return err
			}
		}

		if err = ch.TxCommit(); err != nil {
			err = fmt.Errorf("topology.All: failed to commit topology transaction, %w", err)
		}

		return err
	})
}
