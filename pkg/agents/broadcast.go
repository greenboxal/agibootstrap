package agents

import (
	"context"
	"sync"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type BroadcastRouter struct {
	mu sync.RWMutex

	incomingMessages chan *thoughtstream.Thought
	outgoingMessages []*thoughtstream.Thought

	agentMap map[string]Agent

	log thoughtstream.Branch
}

func NewRouter(log thoughtstream.Branch) *BroadcastRouter {
	return &BroadcastRouter{
		log:              log,
		agentMap:         map[string]Agent{},
		incomingMessages: make(chan *thoughtstream.Thought, 32),
	}
}

func (r *BroadcastRouter) OutgoingMessages() []*thoughtstream.Thought { return r.outgoingMessages }

func (r *BroadcastRouter) RouteMessage(ctx context.Context, msg *thoughtstream.Thought) error {
	return r.routeMessage(ctx, msg)
}

func (r *BroadcastRouter) RegisterAgent(agent Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.agentMap[agent.Profile().Name] = agent

	agent.AttachTo(r)
}

func (r *BroadcastRouter) ReceiveIncomingMessage(msg *thoughtstream.Thought) {
	r.incomingMessages <- msg
}

func (r *BroadcastRouter) RouteIncomingMessages(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-r.incomingMessages:
			if err := r.routeMessage(ctx, msg); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func (r *BroadcastRouter) routeMessage(ctx context.Context, msg *thoughtstream.Thought) error {
	if r.log != nil {
		r.log.Mutate().Append(msg)
	}

	if msg.ReplyTo != nil {
		if msg.ReplyTo.Role == msn.RoleUser {
			r.outgoingMessages = append(r.outgoingMessages, msg)
		} else {
			agent := r.agentMap[msg.ReplyTo.Name]

			if err := agent.ReceiveMessage(ctx, msg); err != nil {
				return err
			}
		}
	} else {
		for _, agent := range r.agentMap {
			if agent.Profile().Name == msg.From.Name {
				continue
			}

			if err := agent.ReceiveMessage(ctx, msg); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *BroadcastRouter) ResetOutbox() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.outgoingMessages = r.outgoingMessages[0:0]
}
