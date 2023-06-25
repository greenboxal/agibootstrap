package singularity

import (
	"context"
	"sync"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type Router struct {
	mu sync.RWMutex

	incomingMessages chan thoughtstream.Thought
	outgoingMessages []thoughtstream.Thought

	agentMap map[string]*Agent
}

func NewRouter() *Router {
	return &Router{
		agentMap:         map[string]*Agent{},
		incomingMessages: make(chan thoughtstream.Thought, 32),
	}
}

func (r *Router) RouteMessage(ctx context.Context, msg thoughtstream.Thought) error {
	return r.routeMessage(msg)
}

func (r *Router) OutgoingMessages() []thoughtstream.Thought {
	return r.outgoingMessages
}

func (r *Router) RegisterAgent(agent *Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.agentMap[agent.Profile().Name] = agent

	agent.AttachTo(r)
}

func (r *Router) ReceiveIncomingMessage(msg thoughtstream.Thought) {
	r.incomingMessages <- msg
}

func (r *Router) RouteIncomingMessages(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-r.incomingMessages:
			if err := r.routeMessage(msg); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func (r *Router) routeMessage(msg thoughtstream.Thought) error {
	if msg.ReplyTo != nil {
		if msg.ReplyTo.Role == msn.RoleUser {
			r.outgoingMessages = append(r.outgoingMessages, msg)
		} else {
			agent := r.agentMap[msg.ReplyTo.Name]

			if err := agent.ReceiveMessage(msg); err != nil {
				return err
			}
		}
	} else {
		for _, agent := range r.agentMap {
			if agent.Profile().Name == msg.From.Name {
				continue
			}

			if err := agent.ReceiveMessage(msg); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Router) ResetOutbox() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.outgoingMessages = r.outgoingMessages[0:0]
}
