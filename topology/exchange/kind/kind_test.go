package kind_test

import (
	"testing"

	"github.com/ar3s3ru/go-carrot/topology/exchange/kind"

	"github.com/stretchr/testify/assert"
)

func TestKind_String(t *testing.T) {
	testcases := []struct {
		input    kind.Kind
		expected string
	}{
		{input: kind.Direct, expected: "direct"},
		{input: kind.Fanout, expected: "fanout"},
		{input: kind.Headers, expected: "headers"},
		{input: kind.Internal, expected: "internal"},
		{input: kind.Topic, expected: "topic"},
		{input: kind.Kind("unknown"), expected: "unknown"},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.expected, func(t *testing.T) { assert.Equal(t, tc.expected, tc.input.String()) })
	}
}
