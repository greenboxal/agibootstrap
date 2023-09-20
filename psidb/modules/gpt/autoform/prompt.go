package autoform

import (
	"context"
	"io"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/gpt/promptml"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	gpt "github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type PromptBuilder struct {
	client       *openai.Client
	modelOptions gpt.ModelOptions

	tokenizer tokenizers.BasicTokenizer

	hooks    map[PromptBuilderHook][]PromptBuilderHookFunc
	messages map[PromptBuilderHook][]PromptMessageSource

	enableTools *bool
	forceTool   *string
	tools       map[string]PromptBuilderTool

	focus       *chat.Message
	allMessages []*chat.Message

	Context map[string]any
}

func NewPromptBuilder() *PromptBuilder {
	b := &PromptBuilder{
		Context: map[string]any{},

		hooks:    map[PromptBuilderHook][]PromptBuilderHookFunc{},
		messages: map[PromptBuilderHook][]PromptMessageSource{},
		tools:    map[string]PromptBuilderTool{},

		tokenizer: gpt.GlobalModelTokenizer,
	}

	return b
}

func (b *PromptBuilder) WithClient(client *openai.Client) { b.client = client }

func (b *PromptBuilder) AllMessages() []*chat.Message { return b.allMessages }

func (b *PromptBuilder) AddHook(hook PromptBuilderHook, fn PromptBuilderHookFunc) {
	b.hooks[hook] = append(b.hooks[hook], fn)
}

func (b *PromptBuilder) AppendMessageSources(hook PromptBuilderHook, srcs ...PromptMessageSource) {
	b.messages[hook] = append(b.messages[hook], srcs...)
}

func (b *PromptBuilder) AppendMessage(hook PromptBuilderHook, msg ...*chat.Message) {
	b.AppendMessageSources(hook, StaticMessageSource(msg...))
}

func (b *PromptBuilder) AppendModelMessage(hook PromptBuilderHook, msg ...openai.ChatCompletionMessage) {
	mapped := lo.Map(msg, func(m openai.ChatCompletionMessage, _ int) *chat.Message {
		msg := chat.NewMessage(chat.MessageKindEmit)
		msg.FromOpenAI(m)
		return msg
	})

	b.AppendMessageSources(hook, func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*chat.Message], error) {
		return iterators.FromSlice(mapped), nil
	})
}

func (b *PromptBuilder) SetFocus(msg *chat.Message) { b.focus = msg }
func (b *PromptBuilder) GetFocus() *chat.Message    { return b.focus }

func (b *PromptBuilder) WithModelOptions(opts gpt.ModelOptions) {
	b.modelOptions = b.modelOptions.MergeWith(opts)
}

func (b *PromptBuilder) DisableTools() {
	t := false
	b.enableTools = &t
}

func (b *PromptBuilder) EnableTools() {
	t := true
	b.enableTools = &t
}

func (b *PromptBuilder) AutoTools() {
	b.enableTools = nil
}

func (b *PromptBuilder) ForceTool(name string) {
	b.forceTool = &name
}

func (b *PromptBuilder) WithTools(tools ...PromptBuilderTool) {
	for _, t := range tools {
		b.tools[t.ToolName()] = t
	}
}

func (b *PromptBuilder) Build(ctx context.Context) openai.ChatCompletionRequest {
	var request openai.ChatCompletionRequest

	if len(b.tools) > 0 {
		for _, tool := range b.tools {
			request.Functions = append(request.Functions, *tool.ToolDefinition())
		}
	}

	if b.forceTool != nil {
		request.FunctionCall = struct {
			Name string `json:"name"`
		}{Name: *b.forceTool}
	} else if b.enableTools != nil {
		if *b.enableTools {
			request.FunctionCall = "auto"
		} else {
			request.FunctionCall = "none"
		}
	}

	b.modelOptions.Apply(&request)

	buildStream := func(hook PromptBuilderHook) promptml.Parent {
		return promptml.NewDynamicList(func(ctx context.Context) iterators.Iterator[promptml.Node] {
			for _, fn := range b.hooks[hook] {
				fn(ctx, b, &request)
			}

			messageSources := b.messages[hook]

			if len(messageSources) == 0 {
				return iterators.Empty[promptml.Node]()
			}

			msgSrcIterator := iterators.FromSlice(messageSources)

			if hook == PromptBuilderHookFocus && b.focus != nil {
				msgSrcIterator = iterators.Concat(msgSrcIterator, iterators.Single(StaticMessageSource(b.focus)))
			}

			return iterators.FlatMap(msgSrcIterator, func(fn PromptMessageSource) iterators.Iterator[promptml.Node] {
				msgs, err := fn(ctx, b)

				if err != nil {
					panic(err)
				}

				msgs = iterators.Filter(msgs, func(msg *chat.Message) bool {
					return msg.From.Role != msn.RoleSystem || (msg.Text != "" || msg.FunctionCall != nil)
				})

				if hook != PromptBuilderHookFocus && b.focus != nil {
					msgs = iterators.Filter(msgs, func(msg *chat.Message) bool {
						if msg == b.focus {
							return false
						}

						return true
					})
				}

				return iterators.Map(msgs, func(msg *chat.Message) promptml.Node {
					var options []promptml.StyleOpt[promptml.Node]

					if msg.From.Role == msn.RoleSystem || msg == b.focus {
						options = append(options, promptml.Fixed())
					}

					return promptml.MessageWithUserData(msg.From.Name, msg.From.Role, promptml.Styled(
						promptml.Text(msg.Text),
						options...,
					), msg)
				})
			})
		})
	}

	pml := promptml.Container(
		buildStream(PromptBuilderHookGlobalSystem),
		buildStream(PromptBuilderHookPreHistory),
		buildStream(PromptBuilderHookHistory),
		buildStream(PromptBuilderHookPostHistory),
		buildStream(PromptBuilderHookPreFocus),
		buildStream(PromptBuilderHookFocus),
		buildStream(PromptBuilderHookPostFocus),
		buildStream(PromptBuilderHookLast),
	)

	if err := b.renderPml(ctx, pml, &request); err != nil {
		panic(err)
	}

	return request
}

func (b *PromptBuilder) Execute(ctx context.Context, options ...ExecuteOption) (*PromptResponse, error) {
	var opts ExecuteOptions
	opts.Client = b.client
	opts.ModelOptions = b.modelOptions
	opts.Apply(options...)

	request := b.Build(ctx)

	trace := gpt.CreateTrace(ctx, request)
	defer trace.End()

	runRequest := func(ctx context.Context, req openai.ChatCompletionRequest) error {
		res, err := opts.Client.CreateChatCompletionStream(ctx, request)

		if err != nil {
			panic(err)
		}

		defer res.Close()

		for {
			chunk, err := res.Recv()

			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return err
			}

			trace.ConsumeOpenAI(chunk)

			for _, choice := range chunk.Choices {
				for _, parser := range opts.StreamingParsers {
					if err := parser.ParseChoiceStreamed(ctx, choice); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}

	var recoveryMessages []*openai.ChatCompletionMessage

	handlerError := func(err error) error {
		var recoverable *RecoverableError

		if errors.As(err, &recoverable) {
			if recoveryMessages == nil {
				recoveryMessages = make([]*openai.ChatCompletionMessage, len(trace.Choices))
			}

			// FIXME: Handle multiple recoverable errors on multiple choices
			for i, choice := range trace.Choices {
				if recoveryMessages[i] == nil {
					request.Messages = append(request.Messages, openai.ChatCompletionMessage{
						Role: "assistant",
					})

					recoveryMessages[i] = &request.Messages[len(request.Messages)-1]
				}

				recoveryMessage := recoveryMessages[i]

				partialMessage := choice.Message

				if partialMessage.FunctionCall != nil {
					if recoveryMessage.FunctionCall == nil {
						recoveryMessage.FunctionCall = &openai.FunctionCall{}
					}

					recoveryMessage.FunctionCall.Name = partialMessage.FunctionCall.Name

					recoveryMessage.FunctionCall.Arguments += partialMessage.FunctionCall.Arguments
					recoveryMessage.FunctionCall.Arguments = recoveryMessage.FunctionCall.Arguments[:recoverable.RecoverablePosition.Offset]
				} else {
					recoveryMessage.Content += partialMessage.Content
					recoveryMessage.Content = recoveryMessage.Content[:recoverable.RecoverablePosition.Offset]
				}
			}

			return nil
		}

		return err
	}

	for {
		err := runRequest(ctx, request)

		if err != nil {
			if err := handlerError(err); err != nil {
				return nil, err
			}

			continue
		}

		pr := &PromptResponse{
			Raw: trace,

			Choices: lo.Map(trace.Choices, func(c openai.ChatCompletionChoice, _ int) PromptResponseChoice {
				recoveredMessage := c.Message

				if recoveryMessages != nil && c.Index < len(recoveryMessages) {
					recoveryMessage := recoveryMessages[c.Index]

					if recoveryMessage != nil {
						if recoveryMessage.FunctionCall != nil {
							if recoveredMessage.FunctionCall == nil {
								recoveredMessage.FunctionCall = &openai.FunctionCall{}
							}

							recoveredMessage.FunctionCall.Name = recoveryMessage.FunctionCall.Name
							recoveredMessage.FunctionCall.Arguments = recoveryMessage.FunctionCall.Arguments + recoveredMessage.FunctionCall.Arguments
						}

						recoveredMessage.Content = recoveryMessage.Content + recoveredMessage.Content
					}
				}

				msg := chat.NewMessage(chat.MessageKindEmit)
				msg.FromOpenAI(recoveredMessage)

				choice := PromptResponseChoice{
					Index:   c.Index,
					Reason:  c.FinishReason,
					Message: msg,
				}

				if msg.FunctionCall != nil {
					choice.Tool = &PromptToolSelection{
						Name:      msg.FunctionCall.Name,
						Arguments: msg.FunctionCall.Arguments,
					}
				}

				return choice
			}),
		}

		for _, choice := range pr.Choices {
			for _, parser := range opts.StreamingParsers {
				if err := parser.ParseChoice(ctx, choice); err != nil {
					if err := handlerError(err); err != nil {
						return nil, err
					}

					continue
				}
			}
		}

		return pr, nil
	}
}

func (b *PromptBuilder) ExecuteAndParse(ctx context.Context, parser ResultParser, options ...ExecuteOption) error {
	if sp, ok := parser.(StreamedResultParser); ok {
		options = append(options, ExecuteWithStreamingParser(sp))
	}

	result, err := b.Execute(ctx, options...)

	if err != nil {
		return err
	}

	for _, choice := range result.Choices {
		if err := parser.ParseChoice(ctx, choice); err != nil {
			return err
		}
	}

	return nil
}

func (b *PromptBuilder) renderPml(ctx context.Context, root promptml.Parent, o *openai.ChatCompletionRequest) (err error) {
	stage := promptml.NewStage(root, gpt.GlobalModelTokenizer)
	stage.MaxTokens = (gpt.GlobalClient.MaxTokensForModel(o.Model) / 4 * 3) - int((float64(o.MaxTokens)*1.1)+10) - 256

	if err := root.Update(ctx); err != nil {
		return err
	}

	if str, err := stage.RenderToString(ctx); err != nil {
		str = str
		return err
	}

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

		text, err := stage.RenderNodeToString(ctx, msg)

		if err != nil {
			return err
		}

		originalMsg, hasOriginalMsg := msg.UserData.(*chat.Message)

		m := openai.ChatCompletionMessage{
			Name:    msg.From.Value(),
			Role:    openai.ConvertFromRole(msg.Role.Value()),
			Content: text,
		}

		if hasOriginalMsg {
			if originalMsg.FunctionCall != nil {
				m.FunctionCall = &openai.FunctionCall{
					Name:      originalMsg.FunctionCall.Name,
					Arguments: originalMsg.FunctionCall.Arguments,
				}
			}
		}

		o.Messages = append(o.Messages, m)

		return nil
	}); err != nil {
		return
	}

	return
}

type ExecuteOptions struct {
	Client       *openai.Client
	ModelOptions gpt.ModelOptions

	StreamingParsers []StreamedResultParser
}

func (o *ExecuteOptions) Apply(options ...ExecuteOption) {
	for _, opt := range options {
		opt(o)
	}
}

type ExecuteOption func(o *ExecuteOptions)

func ExecuteWithStreamingParser(parser StreamedResultParser) ExecuteOption {
	return func(o *ExecuteOptions) {
		o.StreamingParsers = append(o.StreamingParsers, parser)
	}
}

func ExecuteWithClient(client *openai.Client) ExecuteOption {
	return func(o *ExecuteOptions) {
		o.Client = client
	}
}

func ExecuteWithModelOptions(opts gpt.ModelOptions) ExecuteOption {
	return func(o *ExecuteOptions) {
		o.ModelOptions = o.ModelOptions.MergeWith(opts)
	}
}
