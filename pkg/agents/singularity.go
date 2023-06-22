package agents

import (
	"context"
	"fmt"
	"sort"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
)

type Singularity struct {
	self *Agent

	agents        map[string]*Agent
	orderedAgents []*Agent

	sharedSystemMessages []chat.Message

	incomingMessages chan Message
	outgoingMessages []Message

	globalLog []Message
}

func NewSingularity() *Singularity {
	s := &Singularity{
		agents: map[string]*Agent{},

		incomingMessages: make(chan Message, 32),
	}

	for _, profile := range AgentProfiles {
		a := NewAgent(profile)

		s.RegisterAgent(a)
	}

	s.self = s.agents[SingularityProfile.Name]

	return s
}

func (s *Singularity) RegisterAgent(agent *Agent) {
	s.agents[agent.Profile().Name] = agent

	agent.AttachTo(s)

	s.orderedAgents = append(s.orderedAgents, agent)
}

func (s *Singularity) Step(ctx context.Context) ([]Message, error) {
	s.outgoingMessages = nil

	s.RouteIncomingMessages(ctx)

	sort.SliceStable(s.orderedAgents, func(i, j int) bool {
		a := s.orderedAgents[i]
		b := s.orderedAgents[j]

		if b.Profile().Rank > a.profile.Rank {
			return false
		} else if b.Profile().Rank < a.profile.Rank {
			return true
		}

		return b.Profile().Priority > a.profile.Priority
	})

	for _, a := range s.orderedAgents {
		fmt.Printf("#\n\n# AGENT %s\n#\n#\n", a.Profile().Name)

		if err := a.Step(ctx); err != nil {
			return s.outgoingMessages, err
		}
	}

	return s.outgoingMessages, nil
}

func (s *Singularity) ReceiveIncomingMessage(msg Message) {
	s.incomingMessages <- msg
}

func (s *Singularity) RouteIncomingMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.incomingMessages:
			s.routeIncomingMessage(msg)
		default:
			return
		}
	}
}

func (s *Singularity) routeIncomingMessage(msg Message) {
	s.globalLog = append(s.globalLog, msg)

	if msg.ReplyTo != nil {
		if msg.ReplyTo.Role == msn.RoleUser {
			s.outgoingMessages = append(s.outgoingMessages, msg)
		} else {
			agent := s.agents[msg.ReplyTo.Name]

			agent.ReceiveMessage(msg)
		}
	} else {
		for _, agent := range s.agents {
			if agent.Profile().Name == msg.From.Name {
				continue
			}

			agent.ReceiveMessage(msg)
		}
	}
}

func (s *Singularity) routeAgentMessage(sender *Agent, msg Message) {
	s.routeIncomingMessage(msg)
}
