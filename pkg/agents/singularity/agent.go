package singularity

import (
	"context"
	"io"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type agentContext struct {
	agent *Agent
	ctx   context.Context
}

func (c agentContext) Context() context.Context       { return c.ctx }
func (c agentContext) Profile() agents.Profile        { return c.agent.profile }
func (c agentContext) Agent() agents.Agent            { return c.agent }
func (c agentContext) Log() *thoughtstream.ThoughtLog { return c.agent.log }
func (c agentContext) Router() agents.Router          { return c.agent.router }
func (c agentContext) WorldState() agents.WorldState  { return c.agent.worldState }

type Agent struct {
	psi.NodeBase

	profile agents.Profile

	router     *Router
	log        *thoughtstream.ThoughtLog
	worldState *WorldState

	lastSummary featureextractors.Summary
}

func NewAgent(
	profile agents.Profile,
	log *thoughtstream.ThoughtLog,
	worldState *WorldState,
) (*Agent, error) {
	a := &Agent{
		profile:    profile,
		log:        log,
		worldState: worldState,
	}

	a.Init(a, "")

	return a, nil
}

func (a *Agent) Log() *thoughtstream.ThoughtLog    { return a.log }
func (a *Agent) WorldState() agents.WorldState     { return a.worldState }
func (a *Agent) Profile() agents.Profile           { return a.profile }
func (a *Agent) History() []*thoughtstream.Thought { return a.log.Messages() }

func (a *Agent) AttachTo(router *Router) {
	a.router = router
}

func (a *Agent) ReceiveMessage(msg *thoughtstream.Thought) error {
	if err := a.log.Push(msg); err != nil {
		return err
	}

	return nil
}

func (a *Agent) ForkSession() (agents.AnalysisSession, error) {
	return NewAgent(a.profile, a.log.ForkTemporary(), a.worldState)
}

func (a *Agent) EmitMessage(msg *thoughtstream.Thought) error {
	msg.Timestamp = time.Now()

	if msg.From.Name == "" && msg.From.Role == msn.RoleAI {
		msg.From.Name = a.profile.Name
	}

	if err := a.log.Push(msg); err != nil {
		return err
	}

	return a.router.routeMessage(msg)
}

func (a *Agent) Step(ctx context.Context) error {
	prompt := agents.ComposePrompt(
		agents.MapToPrompt(a.worldState.SystemMessages, func(c chat.Message) agents.AgentPrompt {
			return agents.MessageToPrompt(c)
		}),

		agents.LogHistory(),

		agents.AgentMessage(a.profile.Name, " "),
	)

	reply, err := a.Introspect(ctx, prompt)

	if err != nil {
		return err
	}

	if err := a.EmitMessage(reply); err != nil {
		return err
	}

	if a.profile.PostStep != nil {
		if err := a.profile.PostStep(agentContext{agent: a, ctx: ctx}, reply); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) Introspect(ctx context.Context, prompt agents.AgentPrompt) (reply *thoughtstream.Thought, err error) {
	var options []llm.PredictOption

	msg, err := prompt.Render(agentContext{agent: a, ctx: ctx})

	if err != nil {
		return nil, err
	}

	stream, err := gpt.GlobalModel.PredictChatStream(ctx, msg, options...)

	if err != nil {
		return nil, err
	}

	reply = a.log.BeginNext()
	reply.From.Name = a.profile.Name
	reply.From.Role = msn.RoleAI

	for {
		frag, err := stream.Recv()

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return reply, err
		}

		reply.Text += frag.Delta

		reply.Invalidate()
		reply.Update()
	}

	return reply, nil
}

func (a *Agent) RunPostCycleHooks(ctx context.Context) error {
	/*summary, err := featureextractors.QuerySummary(ctx, a.log.Messages())

	if err != nil {
		return err
	}

	a.lastSummary = summary*/

	a.log.EpochBarrier()

	return nil
}
