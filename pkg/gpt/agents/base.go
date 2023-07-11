package agents

import (
	"context"
	"fmt"
	"io"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/gpt/promptml"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type AgentContextBase struct {
	ctx context.Context

	agent Agent

	branch thoughtdb.Branch
	stream thoughtdb.Cursor
}

func (c AgentContextBase) Context() context.Context { return c.ctx }
func (c AgentContextBase) Profile() *Profile        { return c.agent.Profile() }
func (c AgentContextBase) Agent() Agent             { return c.agent }
func (c AgentContextBase) Branch() thoughtdb.Branch { return c.branch }
func (c AgentContextBase) Stream() thoughtdb.Cursor { return c.stream }
func (c AgentContextBase) WorldState() WorldState   { return c.agent.WorldState() }

type AgentBase struct {
	psi.NodeBase

	profile *Profile

	log        thoughtdb.Branch
	repo       *thoughtdb.Repo
	router     Router
	worldState WorldState
}

func (a *AgentBase) Init(
	self Agent,
	profile *Profile,
	repo *thoughtdb.Repo,
	log thoughtdb.Branch,
	worldState WorldState,
) {
	a.profile = profile
	a.log = log
	a.repo = repo
	a.worldState = worldState

	a.NodeBase.Init(self)
}

func (a *AgentBase) Repo() *thoughtdb.Repo  { return a.repo }
func (a *AgentBase) Log() thoughtdb.Branch  { return a.log }
func (a *AgentBase) WorldState() WorldState { return a.worldState }
func (a *AgentBase) Profile() *Profile      { return a.profile }

func (a *AgentBase) History() []*thoughtdb.Thought {
	return iterators.ToSlice[*thoughtdb.Thought](a.log.Cursor().IterateParents())
}

func (a *AgentBase) AttachTo(router Router) {
	a.router = router
}

func (a *AgentBase) ReceiveMessage(ctx context.Context, msg *thoughtdb.Thought) error {
	msg = msg.Clone()
	msg.Pointer = thoughtdb.Pointer{}

	return a.log.Commit(ctx, msg)
}

func (a *AgentBase) ForkSession() (AnalysisSession, error) {
	forked := &AgentBase{}

	forked.Init(forked, a.profile, a.repo, a.log.Fork(), a.worldState)

	return forked, nil
}

func (a *AgentBase) EmitMessage(ctx context.Context, msg *thoughtdb.Thought) error {
	if msg.From.Name == "" && msg.From.Role == msn.RoleAI {
		msg.From.Name = a.profile.Name
	}

	return a.router.RouteMessage(ctx, msg)
}

func (a *AgentBase) Step(ctx context.Context, options ...StepOption) error {
	var opts StepOptions

	opts.Base = a.log
	opts.Head = a.log.Pointer()

	if err := opts.Apply(options...); err != nil {
		return err
	}

	fork := opts.Base.Fork()

	if opts.Prompt == nil {
		opts.Prompt = TmlContainer(
			promptml.Message("", msn.RoleSystem, promptml.Styled(
				promptml.Text(a.profile.BaselineSystemPrompt),
				promptml.Fixed(),
			)),

			promptml.Map(fork.Cursor().IterateParents(), func(thought *thoughtdb.Thought) promptml.AttachableNodeLike {
				return promptml.Message(
					thought.From.Name,
					thought.From.Role,
					promptml.Styled(
						promptml.Text(thought.Text),
						promptml.Fixed(),
					),
				)
			}),

			promptml.Message(a.profile.Name, msn.RoleAI, promptml.Placeholder()),
		)
	}

	options = append(
		options,
		WithStream(fork.Cursor()),
		WithHeadPointer(thoughtdb.Pointer{}),
	)

	reply, err := a.Introspect(ctx, opts.Prompt, options...)

	if err != nil {
		return err
	}

	if err := fork.Commit(ctx, reply); err != nil {
		return err
	}

	if err := a.EmitMessage(ctx, reply); err != nil {
		return err
	}

	if a.profile.PostStep != nil {
		actx := a.makeContext(ctx, opts)

		if err := a.profile.PostStep(actx, reply); err != nil {
			return err
		}
	}

	return opts.Base.Merge(ctx, thoughtdb.FlatTimeMergeStrategy(), fork)
}

func (a *AgentBase) Introspect(ctx context.Context, prompt AgentPrompt, options ...StepOption) (reply *thoughtdb.Thought, err error) {
	var opts StepOptions
	var predictOptions []llm.PredictOption

	opts.Base = a.log
	opts.Head = a.log.Pointer()

	if err := opts.Apply(options...); err != nil {
		return nil, err
	}

	actx := a.makeContext(ctx, opts)

	msg, err := prompt.Render(actx)

	if err != nil {
		return nil, err
	}

	fmt.Printf("\n##############\n%s\n", msg.String())

	replyStream, err := gpt.GlobalModel.PredictChatStream(ctx, msg, predictOptions...)

	if err != nil {
		return nil, err
	}

	reply = thoughtdb.NewThought()
	reply.From.Name = a.profile.Name
	reply.From.Role = msn.RoleAI

	for {
		frag, err := replyStream.Recv()

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return reply, err
		}

		reply.Text += frag.Delta

		fmt.Printf("%s", frag.Delta)

		reply.Invalidate()

		if err := reply.OnUpdate(ctx); err != nil {
			return nil, err
		}
	}

	fmt.Printf("\n\n##############\n")

	return reply, nil
}

func (a *AgentBase) RunPostCycleHooks(ctx context.Context) error {

	return nil
}

func (a *AgentBase) makeContext(ctx context.Context, options StepOptions) AgentContext {
	return AgentContextBase{
		ctx:    ctx,
		agent:  a,
		branch: options.Base,
		stream: options.Stream,
	}
}
