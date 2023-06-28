package agents

import (
	"context"
	"io"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type AgentContextBase struct {
	ctx context.Context

	agent Agent
}

func (c AgentContextBase) Context() context.Context  { return c.ctx }
func (c AgentContextBase) Profile() *Profile         { return c.agent.Profile() }
func (c AgentContextBase) Agent() Agent              { return c.agent }
func (c AgentContextBase) Log() thoughtstream.Branch { return c.agent.Log() }
func (c AgentContextBase) WorldState() WorldState    { return c.agent.WorldState() }

type AgentBase struct {
	psi.NodeBase

	profile *Profile

	log        thoughtstream.Branch
	router     Router
	worldState WorldState
}

func (a *AgentBase) Init(
	self Agent,
	profile *Profile,
	log thoughtstream.Branch,
	worldState WorldState,
) {
	a.profile = profile
	a.log = log
	a.worldState = worldState

	a.NodeBase.Init(self, "")
}

func (a *AgentBase) Log() thoughtstream.Branch { return a.log }
func (a *AgentBase) WorldState() WorldState    { return a.worldState }
func (a *AgentBase) Profile() *Profile         { return a.profile }

func (a *AgentBase) History() []*thoughtstream.Thought {
	return iterators.ToSlice[*thoughtstream.Thought](a.log.Stream())
}

func (a *AgentBase) AttachTo(router Router) {
	a.router = router
}

func (a *AgentBase) ReceiveMessage(ctx context.Context, msg *thoughtstream.Thought) error {
	a.log.Mutate().Append(msg)

	return nil
}

func (a *AgentBase) ForkSession() (AnalysisSession, error) {
	forked := &AgentBase{}

	forked.Init(forked, a.profile, a.log.Stream().Fork().AsBranch(), a.worldState)

	return forked, nil
}

func (a *AgentBase) EmitMessage(ctx context.Context, msg *thoughtstream.Thought) error {
	msg.Pointer.Timestamp = time.Now()

	if msg.From.Name == "" && msg.From.Role == msn.RoleAI {
		msg.From.Name = a.profile.Name
	}

	a.log.Mutate().Append(msg)

	return a.router.RouteMessage(ctx, msg)
}

func (a *AgentBase) Step(ctx context.Context) error {
	prompt := ComposePrompt(
		ThoughtHistory(a.log.Stream()),
		AgentMessage(a.profile.Name, " "),
	)

	reply, err := a.Introspect(ctx, prompt)

	if err != nil {
		return err
	}

	if err := a.EmitMessage(ctx, reply); err != nil {
		return err
	}

	if a.profile.PostStep != nil {
		if err := a.profile.PostStep(AgentContextBase{agent: a, ctx: ctx}, reply); err != nil {
			return err
		}
	}

	return nil
}

func (a *AgentBase) Introspect(ctx context.Context, prompt AgentPrompt) (reply *thoughtstream.Thought, err error) {
	var options []llm.PredictOption

	msg, err := prompt.Render(AgentContextBase{agent: a, ctx: ctx})

	if err != nil {
		return nil, err
	}

	stream, err := gpt.GlobalModel.PredictChatStream(ctx, msg, options...)

	if err != nil {
		return nil, err
	}

	reply = thoughtstream.NewThought()
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

func (a *AgentBase) RunPostCycleHooks(ctx context.Context) error {

	return nil
}
