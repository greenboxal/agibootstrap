package singularity

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	agents "github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type Singularity struct {
	mu   sync.RWMutex
	self *Agent
	log  *thoughtstream.AgentLog

	sharedSystemMessages []chat.Message

	incomingMessages chan thoughtstream.Thought
	outgoingMessages []thoughtstream.Thought

	agents        []*Agent
	agentMap      map[string]*Agent
	agentStateMap map[string]*AgentState

	worldState *WorldState
}

type AgentState struct {
	Agent *Agent

	Attention float64
}

func (as *AgentState) AttentionScore() float64 {
	p := as.Agent.Profile()

	attnPriority := p.Priority * (1.0 + 1.0/as.Attention)
	ranked := (p.Rank * 1000) * (1.0 + 1.0/as.Attention)

	inv := math.Sqrt(attnPriority*attnPriority + ranked*ranked)

	return 1.0 / inv
}

func NewSingularity(lm *thoughtstream.Manager) *Singularity {
	s := &Singularity{
		agentMap:      map[string]*Agent{},
		agentStateMap: map[string]*AgentState{},

		incomingMessages: make(chan thoughtstream.Thought, 32),

		log: thoughtstream.NewAgentLog("GLOBAL"),

		worldState: NewWorldState(),
	}

	for _, profile := range AgentProfiles {
		a, err := NewAgent(lm, profile)

		if err != nil {
			panic(err)
		}

		s.RegisterAgent(a)
	}

	s.self = s.agentMap[SingularityProfile.Name]

	return s
}

func (s *Singularity) WorldState() agents.WorldState {
	return s.worldState
}

func (s *Singularity) AgentState(agent *Agent) *AgentState {
	if att := s.agentStateMap[agent.Profile().Name]; att != nil {
		return att
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	att := s.agentStateMap[agent.Profile().Name]

	if att == nil {
		att = &AgentState{
			Agent:     agent,
			Attention: 1.0,
		}

		s.agentStateMap[agent.Profile().Name] = att
	}

	return att
}

func (s *Singularity) RequestAttention(agent *Agent, attention float64) {
	s.AgentState(agent).Attention += attention
}

func (s *Singularity) RegisterAgent(agent *Agent) {
	s.agentMap[agent.Profile().Name] = agent

	agent.AttachTo(s)

	s.agents = append(s.agents, agent)
}

func (s *Singularity) Step(ctx context.Context) ([]thoughtstream.Thought, error) {
	s.worldState.Cycle++

	s.outgoingMessages = nil

	if err := s.RouteIncomingMessages(ctx); err != nil {
		return s.outgoingMessages, err
	}

	sort.SliceStable(s.agents, func(i, j int) bool {
		a := s.agents[i]
		b := s.agents[j]
		sa := s.AgentState(a)
		sb := s.AgentState(b)

		return sa.AttentionScore() < sb.AttentionScore()
	})

	availableMsg := "Available agents:\n"

	for _, a := range s.agents {
		sa := s.AgentState(a)

		sa.Attention = 1.0

		availableMsg += fmt.Sprintf("  - **%s:** %s\n", a.Profile().Name, a.Profile().Description)
	}

	for _, a := range s.agents {
		s.worldState.Step++

		kvJson, err := json.Marshal(s.worldState.KV)

		if err != nil {
			return s.outgoingMessages, err
		}

		s.sharedSystemMessages = []chat.Message{
			chat.Compose(chat.Entry(msn.RoleSystem, fmt.Sprintf(`
===
**System Epoch:** %d:%d.%d
**System Clock:** %s
**Global State:**
`+"```json"+`
%s
`+"```"+`
===
`, s.worldState.Epoch, s.worldState.Cycle, s.worldState.Step, time.Now().Format("2006/01/02 - 15:04:05"), kvJson))),

			chat.Compose(chat.Entry(msn.RoleSystem, availableMsg)),
		}

		fmt.Printf("#\n#\n# AGENT %s\n#\n#\n", a.Profile().Name)

		if err := a.Step(ctx); err != nil {
			return s.outgoingMessages, err
		}

		data, err := json.Marshal(s.worldState)

		if err != nil {
			return s.outgoingMessages, err
		}

		err = os.WriteFile("/tmp/agib-agent-logs/state.json", data, 0644)

		if err != nil {
			return s.outgoingMessages, err
		}
	}

	var wg sync.WaitGroup

	for _, a := range s.agents {
		wg.Add(1)

		go func(a *Agent) {
			if err := a.RunPostCycleHooks(ctx); err != nil {
				fmt.Printf("Error running post-cycle hooks for %s: %s\n", a.Profile().Name, err.Error())
			}
		}(a)
	}

	wg.Wait()

	return s.outgoingMessages, nil
}

func (s *Singularity) ReceiveIncomingMessage(msg thoughtstream.Thought) {
	s.incomingMessages <- msg
}

func (s *Singularity) RouteIncomingMessages(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-s.incomingMessages:
			if err := s.routeIncomingMessage(msg); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func (s *Singularity) routeIncomingMessage(msg thoughtstream.Thought) error {
	s.log.Message(msg)

	if msg.ReplyTo != nil {
		if msg.ReplyTo.Role == msn.RoleUser {
			s.outgoingMessages = append(s.outgoingMessages, msg)
		} else {
			agent := s.agentMap[msg.ReplyTo.Name]

			if err := agent.ReceiveMessage(msg); err != nil {
				return err
			}
		}
	} else {
		for _, agent := range s.agentMap {
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

func (s *Singularity) routeAgentMessage(sender *Agent, msg thoughtstream.Thought) error {
	return s.routeIncomingMessage(msg)
}
