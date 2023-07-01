package agents

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type colabAgentContext struct {
	colab *Colab
	ctx   context.Context
}

func (c colabAgentContext) Context() context.Context { return c.ctx }
func (c colabAgentContext) Profile() *Profile        { return c.colab.members[0].Profile() }
func (c colabAgentContext) Agent() Agent             { return c.colab }
func (c colabAgentContext) Branch() thoughtdb.Branch { return c.colab.log }
func (c colabAgentContext) Stream() thoughtdb.Cursor { return c.colab.log.Cursor() }
func (c colabAgentContext) WorldState() WorldState   { return c.colab.state }

type Colab struct {
	psi.NodeBase

	log       thoughtdb.Branch
	router    Router
	state     WorldState
	scheduler Scheduler

	members []Agent
	agents  map[string]Agent
}

func (c *Colab) Members() []Agent       { return c.members }
func (c *Colab) Router() Router         { return c.router }
func (c *Colab) Profile() *Profile      { return c.members[0].Profile() }
func (c *Colab) Log() thoughtdb.Branch  { return c.log }
func (c *Colab) WorldState() WorldState { return c.state }

func (c *Colab) History() []*thoughtdb.Thought {
	return iterators.ToSlice[*thoughtdb.Thought](c.log.Cursor().IterateParents())
}

func (c *Colab) AttachTo(r Router) {
	c.router = r
}

func (c *Colab) ReceiveMessage(ctx context.Context, msg *thoughtdb.Thought) error {
	if err := c.router.RouteMessage(ctx, msg); err != nil {
		return err
	}

	return nil
}

func (c *Colab) Introspect(ctx context.Context, prompt AgentPrompt, options ...StepOption) (*thoughtdb.Thought, error) {
	next, err := c.nextSpeaker(ctx)

	if err != nil {
		return nil, err
	}

	return next.Introspect(ctx, prompt, options...)
}

func (c *Colab) IntrospectWith(ctx context.Context, profileName string, prompt AgentPrompt, options ...StepOption) (*thoughtdb.Thought, error) {
	next := c.agents[profileName]

	if next == nil {
		return nil, errors.New("no such agent")
	}

	return next.Introspect(ctx, prompt, options...)
}

func (c *Colab) StepWith(ctx context.Context, profileName string, options ...StepOption) error {
	next := c.agents[profileName]

	if next == nil {
		return errors.New("no such agent")
	}

	return next.Step(ctx, options...)
}

func (c *Colab) Step(ctx context.Context, options ...StepOption) error {
	if err := c.router.RouteIncomingMessages(ctx); err != nil {
		return err
	}

	next, err := c.nextSpeaker(ctx)

	if err != nil {
		return err
	}

	return next.Step(ctx, options...)
}

func (c *Colab) ForkSession() (AnalysisSession, error) {
	return NewColab(c.state, c.log.Fork(), c.scheduler, c.members[0], c.members[1:]...)
}

func (c *Colab) nextSpeaker(ctx context.Context) (Agent, error) {
	return c.scheduler.NextSpeaker(colabAgentContext{ctx: ctx, colab: c}, maps.Values(c.agents)...)
}

func NewColab(state WorldState, log thoughtdb.Branch, scheduler Scheduler, leader Agent, members ...Agent) (*Colab, error) {
	c := &Colab{
		log:       log,
		scheduler: scheduler,
		members:   append([]Agent{leader}, members...),
		agents:    map[string]Agent{},
		state:     state,
	}

	c.Init(c, "")

	c.router = NewBroadcastRouter(c.log)

	for _, member := range c.members {
		member.SetParent(c)

		c.router.RegisterAgent(member)

		c.agents[member.Profile().Name] = member
	}

	return c, nil
}
