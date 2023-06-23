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
)

type Agent struct {
	profile agents.Profile

	s   *Singularity
	log *agents.ChatLog

	lastSummary featureextractors.Summary
}

func (a *Agent) ForkSession() agents.AnalysisSession {
	return &Agent{
		profile: a.profile,
		log:     a.log.ForkTemporary(),
		s:       a.s,
	}
}

func NewAgent(profile agents.Profile) (*Agent, error) {
	log, err := agents.NewChatLog(profile.Name)

	if err != nil {
		return nil, err
	}

	a := &Agent{
		profile: profile,
		log:     log,
	}

	return a, nil
}

func (a *Agent) Profile() agents.Profile   { return a.profile }
func (a *Agent) History() []agents.Message { return a.log.Messages() }

func (a *Agent) AttachTo(s *Singularity) {
	a.s = s
}

func (a *Agent) ReceiveMessage(msg agents.Message) error {
	if err := a.log.Push(msg); err != nil {
		return err
	}

	return nil
}

func (a *Agent) EmitMessage(msg agents.Message) error {
	msg.Timestamp = time.Now()

	if msg.From.Name == "" && msg.From.Role == msn.RoleAI {
		msg.From.Name = a.profile.Name
	}

	if err := a.log.Push(msg); err != nil {
		return err
	}

	return a.s.routeAgentMessage(a, msg)
}

func (a *Agent) Step(ctx context.Context) error {
	reply, err := a.Introspect(ctx)

	if err != nil {
		return err
	}

	for _, entry := range reply.Entries {
		msg := agents.Message{
			From: agents.CommHandle{
				Name: entry.Name,
				Role: entry.Role,
			},

			Text: entry.Text,
		}

		if err := a.EmitMessage(msg); err != nil {
			return err
		}
	}

	if a.profile.PostStep != nil {
		if err := a.profile.PostStep(ctx, a, reply, a.s.worldState); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) Introspect(ctx context.Context, extra ...agents.Message) (chat.Message, error) {
	logMessages := a.log.Messages()

	if len(extra) > 0 {
		msgs := make([]agents.Message, 0, len(logMessages))
		msgs = append(msgs, logMessages...)
		msgs = append(msgs, extra...)
		logMessages = msgs
	}

	prompt := &agents.AgentPromptTemplate{
		Profile:        a.profile,
		SystemMessages: a.s.sharedSystemMessages,
		Messages:       logMessages,
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
	summary, err := featureextractors.QuerySummary(ctx, a.log.Messages())

	if err != nil {
		return err
	}

	a.lastSummary = summary

	return nil
}
