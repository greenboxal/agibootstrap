package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type ChatHistory interface {
	ForkAsChatLog(ctx context.Context, baseMessage *chat.Message, options gpt.ModelOptions) (ChatLog, error)

	MessageIterator(ctx context.Context) (iterators.Iterator[*chat.Message], error)
}

type ChatLog interface {
	ChatHistory

	AcceptMessage(ctx context.Context, msg *chat.Message) error
	AcceptChoice(ctx context.Context, baseMessage *chat.Message, choice PromptResponseChoice) error
}

func MessageSourceFromChatHistory(history ChatHistory) PromptMessageSource {
	return func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*chat.Message], error) {
		return history.MessageIterator(ctx)
	}
}

type ChatHistoryFunc func(ctx context.Context) (iterators.Iterator[*chat.Message], error)

func (c ChatHistoryFunc) Messages(ctx context.Context) (iterators.Iterator[*chat.Message], error) {
	return c(ctx)
}

func (c *Conversation) MessageIterator(ctx context.Context) (iterators.Iterator[*chat.Message], error) {
	msgs, err := c.SliceMessages(ctx, nil, nil)

	if err != nil {
		return nil, err
	}

	return iterators.FromSlice(msgs), nil
}
