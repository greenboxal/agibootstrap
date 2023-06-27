package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type AnalysisSession interface {
	psi.Node

	History() []*thoughtstream.Thought
	Introspect(ctx context.Context, prompt AgentPrompt) (*thoughtstream.Thought, error)

	ReceiveMessage(ctx context.Context, msg *thoughtstream.Thought) error
}
