package topology

import (
	"errors"
	"testing"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	t.Run("no declarers, nothing done, no error returned", func(t *testing.T) {
		assert.NoError(t, All().Declare(nil))
	})

	t.Run("all declarer are called if none fails", func(t *testing.T) {
		var called1 bool
		var called2 bool

		declarer := All(
			declarerFunc(func(*amqp.Channel) error {
				called1 = true
				return nil
			}),
			declarerFunc(func(*amqp.Channel) error {
				called2 = true
				return nil
			}),
		)

		assert.NoError(t, declarer.Declare(nil))
		assert.True(t, called1)
		assert.True(t, called2)
	})

	t.Run("fails with first declarer failed error", func(t *testing.T) {
		var called1 bool
		var called2 bool

		declarer := All(
			declarerFunc(func(*amqp.Channel) error {
				called1 = true
				return errors.New("failed")
			}),
			declarerFunc(func(*amqp.Channel) error {
				called2 = true
				return nil
			}),
		)

		assert.Error(t, declarer.Declare(nil))
		assert.True(t, called1)
		assert.False(t, called2)
	})
}
