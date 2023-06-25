package singularity

import (
	"context"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

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
	reply, err := a.Introspect(ctx)

	if err != nil {
		return err
	}

	for _, entry := range reply.Entries {
		msg := thoughtstream.NewThought()

		msg.From = thoughtstream.CommHandle{
			Name: entry.Name,
			Role: entry.Role,
		}

		msg.Text = entry.Text

		if err := a.EmitMessage(msg); err != nil {
			return err
		}
	}

	if a.profile.PostStep != nil {
		if err := a.profile.PostStep(agentContext{agent: a, ctx: ctx}, reply); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) Introspect(ctx context.Context, extra ...*thoughtstream.Thought) (chat.Message, error) {
	logMessages := a.log.Messages()

	if len(extra) > 0 {
		msgs := make([]*thoughtstream.Thought, 0, len(logMessages))
		msgs = append(msgs, logMessages...)
		msgs = append(msgs, extra...)
		logMessages = msgs
	}

	prompt := &agents.AgentPromptTemplate{
		Profile:        a.profile,
		SystemMessages: a.worldState.SystemMessages,

		Messages: logMessages,
	}

	cctx := chain.NewChainContext(ctx)

	stepChain := chain.New(
		chain.WithName("AgentStep"),

		chain.Sequential(
			chat.Predict(
				gpt.GlobalModel,
				prompt,
				chat.WithMaxTokens(int(2048*(1-a.Profile().Rank))),
			),
		),
	)

	if err := stepChain.Run(cctx); err != nil {
		return chat.Message{}, err
	}

	return chain.Output(cctx, chat.ChatReplyContextKey), nil
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
