package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
)

type Router interface {
	RegisterAgent(agent Agent)

	ReceiveIncomingMessage(ctx context.Context, msg *thoughtdb.Thought)
	RouteIncomingMessages(ctx context.Context) error

	RouteMessage(ctx context.Context, msg *thoughtdb.Thought) error

	OutgoingMessages() []*thoughtdb.Thought
	ResetOutbox()
}
