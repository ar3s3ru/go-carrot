package topology

import (
	"errors"
	"testing"

	"github.com/ar3s3ru/go-carrot/topology/mocks"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	t.Run("no declarers in the combinator returns nil", func(t *testing.T) {
		assert.Nil(t, All())
	})

	t.Run("all declarer are called if none fails", func(t *testing.T) {
		var called1 bool
		var called2 bool

		declarer := All(
			declarerFunc(func(Channel) error {
				called1 = true
				return nil
			}),
			declarerFunc(func(Channel) error {
				called2 = true
				return nil
			}),
		)

		ch := new(mocks.Channel)
		ch.On("Tx").Return(nil).Once()
		ch.On("TxCommit").Return(nil).Once()

		assert.NoError(t, declarer.Declare(ch))
		assert.True(t, called1)
		assert.True(t, called2)
	})

	t.Run("fails with first declarer failed error", func(t *testing.T) {
		var called1 bool
		var called2 bool

		declarer := All(
			declarerFunc(func(Channel) error {
				called1 = true
				return errors.New("failed")
			}),
			declarerFunc(func(Channel) error {
				called2 = true
				return nil
			}),
		)

		ch := new(mocks.Channel)
		ch.On("Tx").Return(nil).Once()
		ch.On("TxRollback").Return(nil).Once()

		assert.Error(t, declarer.Declare(ch))
		assert.True(t, called1)
		assert.False(t, called2)
	})
}
