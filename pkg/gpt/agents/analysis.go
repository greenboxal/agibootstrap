package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type AnalysisSession interface {
	psi.Node

	Introspect(ctx context.Context, prompt AgentPrompt, options ...StepOption) (*thoughtdb.Thought, error)

	ReceiveMessage(ctx context.Context, msg *thoughtdb.Thought) error
}
