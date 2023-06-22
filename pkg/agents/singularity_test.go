package agents

import (
	"context"
	"testing"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/stretchr/testify/require"
)

func TestSingularity(t *testing.T) {
	s := NewSingularity()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.ReceiveIncomingMessage(Message{
		From: CommHandle{
			Name: "Human",
			Role: msn.RoleUser,
		},

		Text: `
Write a Python class that can calculate sequence of numbers like the Fibonacci sequence, among others.
`,
	})

	for {
		msgs, err := s.Step(ctx)

		require.NoError(t, err)

		if len(msgs) > 0 {
			t.Logf("Singularity: %d messages", len(msgs))
		}
	}
}
