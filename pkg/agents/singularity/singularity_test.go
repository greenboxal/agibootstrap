package singularity

import (
	"context"
	"testing"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/pkg/agents"
)

func TestSingularity(t *testing.T) {
	s := NewSingularity()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.ReceiveIncomingMessage(agents.Message{
		Timestamp: time.Now(),

		From: agents.CommHandle{
			Name: "Human",
			Role: msn.RoleUser,
		},

		Text: `
Create a Pytorch model based on the human brain cytoarchitecture.
`,
	})

	st := s.worldState

	for {
		t.Logf("Singularity: Step (epoch = %d, cycle = %d, step = %d)", st.Epoch, st.Cycle, st.Step)
		msgs, err := s.Step(ctx)

		require.NoError(t, err)

		if len(msgs) > 0 {
			t.Logf("Singularity: %d messages", len(msgs))
		}

		progress := agents.GetState(st, CtxGoalStatus)

		if progress.Completed {
			break
		}

		t.Logf("Singularity: Goal progress: %v", progress)
	}
}
