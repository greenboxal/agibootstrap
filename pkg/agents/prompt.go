package agents

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/gpt/promptml"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type AgentPrompt interface {
	Render(ctx AgentContext) (chat.Message, error)
}

type AgentPromptFunc func(ctx AgentContext) (chat.Message, error)

func (a AgentPromptFunc) Render(ctx AgentContext) (chat.Message, error) { return a(ctx) }

func TmlContainer(children ...promptml.Node) AgentPromptFunc {
	return Tml(func(ctx AgentContext) promptml.Parent {
		return promptml.Container(children...)
	})
}

func Tml(rootBuilder func(ctx AgentContext) promptml.Parent) AgentPromptFunc {
	return func(ctx AgentContext) (result chat.Message, err error) {
		root := rootBuilder(ctx)
		stage := promptml.NewStage(root, gpt.GlobalModelTokenizer)
		stage.MaxTokens = 10240

		for it := root.ChildrenIterator(); it.Next(); {
			msg, ok := it.Value().(*promptml.ChatMessage)

			if !ok {
				continue
			}

			text, err := stage.RenderToString(ctx.Context())

			if err != nil {
				return result, err
			}

			m := chat.MessageEntry{
				Role: msg.Role.Value(),
				Name: msg.From.Value(),
				Text: text,
			}

			result.Entries = append(result.Entries, m)
		}

		return
	}
}

func SystemMessage(msg string) AgentPromptFunc {
	return func(ctx AgentContext) (chat.Message, error) {
		return chat.Compose(
			chat.Entry(msn.RoleSystem, msg),
		), nil
	}
}

func AgentMessage(name, text string) AgentPromptFunc {
	return func(ctx AgentContext) (chat.Message, error) {
		return chat.Compose(
			chat.MessageEntry{
				Role: msn.RoleAI,
				Name: name,
				Text: text,
			},
		), nil
	}
}

func ThoughtHistory(iterator thoughtstream.Iterable) AgentPromptFunc {
	return func(ctx AgentContext) (chat.Message, error) {
		msgs := iterators.Map(iterator.Iterator(), func(thought *thoughtstream.Thought) chat.Message {
			return chat.Compose(
				chat.MessageEntry{
					Role: thought.From.Role,
					Name: thought.From.Name,
					Text: thought.Text,
				},
			)
		})

		return chat.Merge(iterators.ToSlice(msgs)...), nil
	}
}

func LogHistory() AgentPromptFunc {
	return func(ctx AgentContext) (chat.Message, error) {
		return ThoughtHistory(ctx.Branch())(ctx)
	}
}

func ComposePrompt(prompts ...AgentPrompt) AgentPromptFunc {
	return func(ctx AgentContext) (chat.Message, error) {
		var result chat.Message

		for _, prompt := range prompts {
			msg, err := prompt.Render(ctx)

			if err != nil {
				return chat.Message{}, err
			}

			result = chat.Merge(result, msg)
		}

		return result, nil
	}
}

func MapToPrompt[T any](items []T, fn func(T) AgentPrompt) AgentPromptFunc {
	return func(ctx AgentContext) (chat.Message, error) {
		var msgs []chat.Message

		for _, item := range items {
			msg, err := fn(item).Render(ctx)

			if err != nil {
				return chat.Message{}, err
			}

			msgs = append(msgs, msg)
		}

		return chat.Merge(msgs...), nil
	}
}

func MessageToPrompt(msg chat.Message) AgentPromptFunc {
	return func(ctx AgentContext) (chat.Message, error) {
		return msg, nil
	}
}
