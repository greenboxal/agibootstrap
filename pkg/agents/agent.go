package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type AgentContext interface {
	Context() context.Context
	Profile() *Profile
	Agent() Agent
	Log() *thoughtstream.ThoughtLog
	WorldState() WorldState
}

type Agent interface {
	AnalysisSession

	Profile() *Profile
	Log() *thoughtstream.ThoughtLog
	WorldState() WorldState

	AttachTo(r Router)
	ForkSession() (AnalysisSession, error)

	Step(ctx context.Context) error
}
