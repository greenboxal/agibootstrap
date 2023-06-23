package agents

import (
	"context"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
)

type PostStepHook func(ctx context.Context, a Agent, msg chat.Message, state WorldState) error

type Profile struct {
	Name        string
	Description any

	BaselineSystemPrompt string

	Rank     float64
	Priority float64

	Provides []string
	Requires []string

	PostStep PostStepHook
}

type CommHandle struct {
	ID   string
	Name string
	Role msn.Role
}

type Message struct {
	Timestamp time.Time
	From      CommHandle
	Text      string

	ReplyTo *CommHandle
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
	History() []Message
	Introspect(ctx context.Context, extra ...Message) (chat.Message, error)
}

type Agent interface {
	AnalysisSession

	Step(ctx context.Context) error

	ForkSession() AnalysisSession
}

type Router interface {
	RouteMessage(ctx context.Context, msg Message) error
}

type Scheduler interface {
	NextStep() (int, error)
}
