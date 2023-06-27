package singularity

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type colabAgentContext struct {
	colab *Colab
	ctx   context.Context
}

func (c colabAgentContext) Context() context.Context       { return c.ctx }
func (c colabAgentContext) Profile() agents.Profile        { return c.colab.members[0] }
func (c colabAgentContext) Agent() agents.Agent            { return c.colab }
func (c colabAgentContext) Log() *thoughtstream.ThoughtLog { return c.colab.log }
func (c colabAgentContext) Router() agents.Router          { return c.colab.router }
func (c colabAgentContext) WorldState() agents.WorldState  { return c.colab.state }

type Colab struct {
	psi.NodeBase

	log       *thoughtstream.ThoughtLog
	router    *Router
	state     *WorldState
	scheduler Scheduler

	members []agents.Profile
	agents  map[string]agents.Agent
}

func (c *Colab) Members() []agents.Profile         { return c.members }
func (c *Colab) Router() *Router                   { return c.router }
func (c *Colab) Profile() agents.Profile           { return c.members[0] }
func (c *Colab) Log() *thoughtstream.ThoughtLog    { return c.log }
func (c *Colab) WorldState() agents.WorldState     { return c.state }
func (c *Colab) History() []*thoughtstream.Thought { return c.log.Messages() }

func (c *Colab) Introspect(ctx context.Context, prompt agents.AgentPrompt) (*thoughtstream.Thought, error) {
	next, err := c.nextSpeaker(ctx)

	if err != nil {
		return nil, err
	}

	return next.Introspect(ctx, prompt)
}

func (c *Colab) IntrospectWith(ctx context.Context, profileName string, prompt agents.AgentPrompt) (*thoughtstream.Thought, error) {
	next := c.agents[profileName]

	if next == nil {
		return nil, errors.New("no such agent")
	}

	return next.Introspect(ctx, prompt)
}

func (c *Colab) StepWith(ctx context.Context, profileName string) error {
	next := c.agents[profileName]

	if next == nil {
		return errors.New("no such agent")
	}

	return next.Step(ctx)
}

func (c *Colab) Step(ctx context.Context) error {
	if err := c.router.RouteIncomingMessages(ctx); err != nil {
		return err
	}

	next, err := c.nextSpeaker(ctx)

	if err != nil {
		return err
	}

	return next.Step(ctx)
}

func (c *Colab) ForkSession() (agents.AnalysisSession, error) {
	return NewColab(c.state, c.log.ForkTemporary(), c.scheduler, c.members[0], c.members[1:]...)
}

func (c *Colab) nextSpeaker(ctx context.Context) (agents.Agent, error) {
	return c.scheduler.NextSpeaker(colabAgentContext{ctx: ctx, colab: c}, maps.Values(c.agents)...)
}

func NewColab(state *WorldState, log *thoughtstream.ThoughtLog, scheduler Scheduler, leader agents.Profile, members ...agents.Profile) (*Colab, error) {
	c := &Colab{
		log:       log,
		scheduler: scheduler,
		members:   append([]agents.Profile{leader}, members...),
		agents:    map[string]agents.Agent{},
		state:     state,
	}

	c.Init(c, "")

	c.router = NewRouter(c.log)

	for _, member := range c.members {
		agent, err := NewAgent(member, c.log.ForkTemporary(), c.state)

		if err != nil {
			return nil, err
		}

		agent.SetParent(c)

		c.router.RegisterAgent(agent)

		c.agents[member.Name] = agent
	}

	return c, nil
}
