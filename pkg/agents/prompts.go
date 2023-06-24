package agents

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
)

type SimplePromptTemplate struct {
	Messages []thoughtstream.Thought
}

func (a *SimplePromptTemplate) AsPrompt() chain.Prompt { panic("implement me") }

func (a *SimplePromptTemplate) Build(ctx chain.ChainContext) (chat.Message, error) {
	messages := make([]chat.MessageEntry, 0, len(a.Messages))

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

type AgentPromptTemplate struct {
	Profile        Profile
	SystemMessages []chat.Message
	Messages       []thoughtstream.Thought
}

func (a *AgentPromptTemplate) AsPrompt() chain.Prompt { panic("implement me") }

func (a *AgentPromptTemplate) Build(ctx chain.ChainContext) (chat.Message, error) {
	extraSystemMessages := chat.Merge(a.SystemMessages...)

	messages := make([]chat.MessageEntry, 0, len(a.Messages)+len(extraSystemMessages.Entries)+2)

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

	messages = append(messages, chat.MessageEntry{
		Role: msn.RoleAI,
		Name: a.Profile.Name,
		Text: " ",
	})

	return chat.Compose(messages...), nil
}
