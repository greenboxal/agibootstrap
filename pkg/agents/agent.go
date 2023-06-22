package agents

import (
	"context"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
)

var oai = openai.NewClient()
var embedder = &openai.Embedder{
	Client: oai,
	Model:  openai.AdaEmbeddingV2,
}
var model = &openai.ChatLanguageModel{
	Client:      oai,
	Model:       "gpt-3.5-turbo-16k",
	Temperature: 1,
}

type Profile struct {
	Name                 string
	BaselineSystemPrompt string

	Rank     float32
	Priority float64
}

type CommHandle struct {
	ID   string
	Name string
	Role msn.Role
}

type Message struct {
	Timestamp time.Time
	From      CommHandle
	Text      string

	ReplyTo *CommHandle
}

type Agent struct {
	profile Profile

	messages []Message

	s *Singularity
}

func NewAgent(profile Profile) *Agent {
	a := &Agent{
		profile: profile,
	}

	return a
}

func (a *Agent) Profile() Profile { return a.profile }

func (a *Agent) ReceiveMessage(msg Message) {
	a.messages = append(a.messages, msg)
}

func (a *Agent) EmitMessage(msg Message) {
	a.messages = append(a.messages, msg)

	a.s.routeAgentMessage(a, msg)
}

func (a *Agent) Step(ctx context.Context) error {
	cctx := chain.NewChainContext(ctx)

	stepChain := chain.New(
		chain.WithName("AgentStep"),

		chain.Sequential(
			chat.Predict(
				model,
				&AgentPromptTemplate{
					Profile:        a.profile,
					SystemMessages: a.s.sharedSystemMessages,
					Messages:       a.messages,
				},
				chat.WithMaxTokens(1024),
			),
		),
	)

	if err := stepChain.Run(cctx); err != nil {
		return err
	}

	reply := chain.Output(cctx, chat.ChatReplyContextKey)

	for _, entry := range reply.Entries {
		a.EmitMessage(Message{
			From: CommHandle{
				Name: entry.Name,
				Role: entry.Role,
			},

			Text: entry.Text,
		})
	}

	return nil
}

func (a *Agent) AttachTo(s *Singularity) {
	a.s = s
}

type AgentPromptTemplate struct {
	Profile        Profile
	SystemMessages []chat.Message
	Messages       []Message
}

func (a *AgentPromptTemplate) AsPrompt() chain.Prompt { panic("implement me") }

func (a *AgentPromptTemplate) Build(ctx chain.ChainContext) (chat.Message, error) {
	extraSystemMessages := chat.Merge(a.SystemMessages...)

	messages := make([]chat.MessageEntry, 0, len(a.Messages)+len(extraSystemMessages.Entries)+1)

	messages = append(messages, chat.MessageEntry{
		Role: msn.RoleSystem,
		Text: a.Profile.BaselineSystemPrompt,
	})

	messages = append(messages, extraSystemMessages.Entries...)

	for _, msg := range a.Messages {
		entry := chat.MessageEntry{
			Name: msg.From.Name,
			Role: msg.From.Role,
			Text: msg.Text,
		}

		messages = append(messages, entry)
	}

	return chat.Compose(messages...), nil
}
