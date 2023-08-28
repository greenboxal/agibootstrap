package agents

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type ChatHistory interface {
	ForkAsChatLog(ctx context.Context, baseMessage *Message, options ModelOptions) (ChatLog, error)

	MessageIterator(ctx context.Context) (iterators.Iterator[*Message], error)
}

type ChatLog interface {
	ChatHistory

	AcceptMessage(ctx context.Context, msg *Message) error
	AcceptChoice(ctx context.Context, baseMessage *Message, choice PromptResponseChoice) error
}

func MessageSourceFromChatHistory(history ChatHistory) PromptMessageSource {
	return func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*Message], error) {
		return history.MessageIterator(ctx)
	}
}

type ChatHistoryFunc func(ctx context.Context) (iterators.Iterator[*Message], error)

func (c ChatHistoryFunc) Messages(ctx context.Context) (iterators.Iterator[*Message], error) {
	return c(ctx)
}
