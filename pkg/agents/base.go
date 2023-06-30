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
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type AgentContextBase struct {
	ctx context.Context

	agent Agent

	branch thoughtstream.Branch
	stream thoughtstream.Stream
}

func (c AgentContextBase) Context() context.Context     { return c.ctx }
func (c AgentContextBase) Profile() *Profile            { return c.agent.Profile() }
func (c AgentContextBase) Agent() Agent                 { return c.agent }
func (c AgentContextBase) Branch() thoughtstream.Branch { return c.branch }
func (c AgentContextBase) Stream() thoughtstream.Stream { return c.stream }
func (c AgentContextBase) WorldState() WorldState       { return c.agent.WorldState() }

type AgentBase struct {
	psi.NodeBase

	profile *Profile

	log        thoughtstream.Branch
	repo       thoughtstream.Resolver
	router     Router
	worldState WorldState
}

func (a *AgentBase) Init(
	self Agent,
	profile *Profile,
	repo thoughtstream.Resolver,
	log thoughtstream.Branch,
	worldState WorldState,
) {
	a.profile = profile
	a.log = log
	a.repo = repo
	a.worldState = worldState

	a.NodeBase.Init(self, "")
}

func (a *AgentBase) Repo() thoughtstream.Resolver { return a.repo }
func (a *AgentBase) Log() thoughtstream.Branch    { return a.log }
func (a *AgentBase) WorldState() WorldState       { return a.worldState }
func (a *AgentBase) Profile() *Profile            { return a.profile }

func (a *AgentBase) History() []*thoughtstream.Thought {
	return iterators.ToSlice[*thoughtstream.Thought](a.log.Stream().Reversed())
}

func (a *AgentBase) AttachTo(router Router) {
	a.router = router
}

func (a *AgentBase) ReceiveMessage(ctx context.Context, msg *thoughtstream.Thought) error {
	msg = msg.Clone()
	msg.Pointer = thoughtstream.Pointer{}

	a.log.Mutate().Append(msg)

	return nil
}

func (a *AgentBase) ForkSession() (AnalysisSession, error) {
	forked := &AgentBase{}

	forked.Init(forked, a.profile, a.repo, a.log.Stream().Fork().AsBranch(), a.worldState)

	return forked, nil
}

func (a *AgentBase) EmitMessage(ctx context.Context, msg *thoughtstream.Thought) error {
	if msg.From.Name == "" && msg.From.Role == msn.RoleAI {
		msg.From.Name = a.profile.Name
	}

	return a.router.RouteMessage(ctx, msg)
}

func (a *AgentBase) Step(ctx context.Context, options ...StepOption) error {
	var opts StepOptions

	opts.Base = a.log
	opts.Head = a.log.HeadPointer()

	if err := opts.Apply(options...); err != nil {
		return err
	}

	branchStream := opts.Base.Fork()

	prompt := TmlContainer(
		promptml.Message("", msn.RoleSystem, promptml.Styled(
			promptml.Text(a.profile.BaselineSystemPrompt),
			promptml.Fixed(),
		)),

		promptml.NewDynamicList(func(ctx context.Context) iterators.Iterator[promptml.Node] {
			src := thoughtstream.NewHierarchicalStream(branchStream.AsBranch()).Reversed()

			return iterators.Map(src, func(thought *thoughtstream.Thought) promptml.Node {
				return promptml.Message(
					thought.From.Name,
					thought.From.Role,
					promptml.Styled(
						promptml.Text(thought.Text),
						promptml.Fixed(),
					),
				)
			})
		}),
	)

	options = append(
		options,
		WithStream(branchStream),
		WithHeadPointer(thoughtstream.Pointer{}),
	)

	reply, err := a.Introspect(ctx, prompt, options...)

	if err != nil {
		return err
	}

	branchStream.Append(reply)

	if err := a.EmitMessage(ctx, reply); err != nil {
		return err
	}

	if a.profile.PostStep != nil {
		actx := a.makeContext(ctx, opts)

		if err := a.profile.PostStep(actx, reply); err != nil {
			return err
		}
	}

	r := a.Repo()

	return opts.Base.Mutate().Merge(
		ctx,
		r,
		thoughtstream.FlatTimeMergeStrategy(),
		branchStream.AsBranch(),
	)
}

func (a *AgentBase) Introspect(ctx context.Context, prompt AgentPrompt, options ...StepOption) (reply *thoughtstream.Thought, err error) {
	var opts StepOptions
	var predictOptions []llm.PredictOption

	opts.Base = a.log
	opts.Head = a.log.HeadPointer()

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

	reply = thoughtstream.NewThought()
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
		reply.Update(nil)
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
