package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type Router interface {
	RegisterAgent(agent Agent)

	RouteIncomingMessages(ctx context.Context) error
	RouteMessage(ctx context.Context, msg *thoughtstream.Thought) error

	OutgoingMessages() []*thoughtstream.Thought
	ResetOutbox()
}
