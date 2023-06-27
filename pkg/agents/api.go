package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type AgentContext interface {
	Context() context.Context
	Profile() Profile
	Agent() Agent
	Log() *thoughtstream.ThoughtLog
	WorldState() WorldState
}

type PostStepHook func(ctx AgentContext, thought *thoughtstream.Thought) error

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

type WorldStateKey[T any] string

func (k WorldStateKey[T]) String() string {
	return string(k)
}

type WorldState interface {
	Get(key string) any
	Set(key string, value any)
}

func GetState[T any](state WorldState, k WorldStateKey[T]) (def T) {
	v := state.Get(k.String())

	if v == nil {
		return def
	}

	return v.(T)
}

func SetState[T any](state WorldState, k WorldStateKey[T], v T) {
	state.Set(k.String(), v)
}

type AnalysisSession interface {
	History() []*thoughtstream.Thought
	Introspect(ctx context.Context, prompt AgentPrompt) (*thoughtstream.Thought, error)
}

type Agent interface {
	AnalysisSession

	Profile() Profile
	Log() *thoughtstream.ThoughtLog
	WorldState() WorldState

	Step(ctx context.Context) error

	ForkSession() (AnalysisSession, error)
}

type Router interface {
	RouteMessage(ctx context.Context, msg *thoughtstream.Thought) error
}

type Scheduler interface {
	NextStep() (int, error)
}
