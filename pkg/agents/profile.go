package agents

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type PostStepHook func(ctx AgentContext, thought *thoughtdb.Thought) error

type Profile struct {
	psi.NodeBase

	Name        string
	Description any

	BaselineSystemPrompt string

	Rank     float64
	Priority float64

	Provides []string
	Requires []string

	PostStep PostStepHook
}
