package agents

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/gpt/promptml"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
)

type AgentPrompt interface {
	Render(ctx AgentContext) (chat.Message, error)
}

type AgentPromptFunc func(ctx AgentContext) (chat.Message, error)

func (a AgentPromptFunc) Render(ctx AgentContext) (_ chat.Message, err error) {
	return a(ctx)
}

func TmlContainer(children ...promptml.AttachableNodeLike) AgentPromptFunc {
	return Tml(func(ctx AgentContext) promptml.Parent {
		return promptml.Container(children...)
	})
}

func Tml(rootBuilder func(ctx AgentContext) promptml.Parent) AgentPromptFunc {
	return func(ctx AgentContext) (result chat.Message, err error) {
		root := rootBuilder(ctx)
		stage := promptml.NewStage(root, gpt.GlobalModelTokenizer)
		stage.MaxTokens = 10240

		if err = psi.Walk(root, func(c psi.Cursor, entering bool) error {
			if !entering {
				return nil
			}

			msg, ok := c.Value().(*promptml.ChatMessage)

			if !ok {
				c.WalkChildren()
				return nil
			} else {
				c.SkipChildren()
			}

			text, err := stage.RenderNodeToString(ctx.Context(), msg)

			if err != nil {
				return err
			}

			m := chat.MessageEntry{
				Role: msg.Role.Value(),
				Name: msg.From.Value(),
				Text: text,
			}

			result.Entries = append(result.Entries, m)

			return nil
		}); err != nil {
			return
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

func ThoughtHistory(iterator thoughtdb.Cursor) AgentPromptFunc {
	return func(ctx AgentContext) (chat.Message, error) {
		msgs := iterators.Map(iterator.IterateParents(), func(thought *thoughtdb.Thought) chat.Message {
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
		return ThoughtHistory(ctx.Branch().Cursor())(ctx)
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
