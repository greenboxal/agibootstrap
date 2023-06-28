package singularity

import (
	"context"
	"testing"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/agents/profiles"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

func TestSingularity(t *testing.T) {
	lm := thoughtstream.NewManager("/tmp/agib-test-log")
	defer lm.Close()

	s, err := NewSingularity(lm)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	request := thoughtstream.NewThought()
	request.From.Name = "Human"
	request.From.Role = msn.RoleUser
	request.Text = `
Create a Pytorch model based on the human brain cytoarchitecture.
`

	s.Router().ReceiveIncomingMessage(ctx, request)

	st := s.worldState

	for {
		t.Logf("Singularity: Step (epoch = %d, cycle = %d, step = %d)", st.Epoch, st.Cycle, st.Step)
		msgs, err := s.Step(ctx)

		require.NoError(t, err)

		if len(msgs) > 0 {
			t.Logf("Singularity: %d messages", len(msgs))
		}

		progress := agents.GetState(st, profiles.CtxGoalStatus)

		if progress.Completed {
			break
		}

		t.Logf("Singularity: Goal progress: %v", progress)
	}
}
