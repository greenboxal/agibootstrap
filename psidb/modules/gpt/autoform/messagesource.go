package autoform

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type PromptMessageSource func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*chat.Message], error)

func MessageSourceFromModelMessages(messages ...openai.ChatCompletionMessage) PromptMessageSource {
	return func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*chat.Message], error) {
		return iterators.Map(iterators.FromSlice(messages), func(msg openai.ChatCompletionMessage) *chat.Message {
			m := &chat.Message{}
			m.FromOpenAI(msg)
			return m
		}), nil
	}
}

func StaticMessageSource(items ...*chat.Message) PromptMessageSource {
	return func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*chat.Message], error) {
		return iterators.FromSlice(items), nil
	}
}

func ConcatMessageSource(items ...PromptMessageSource) PromptMessageSource {
	return func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*chat.Message], error) {
		return iterators.FlatMap(iterators.FromSlice(items), func(t PromptMessageSource) iterators.Iterator[*chat.Message] {
			it, err := t(ctx, pb)

			if err != nil {
				panic(err)
			}

			return it
		}), nil
	}
}
