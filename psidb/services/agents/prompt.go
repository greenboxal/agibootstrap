package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/gpt/promptml"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	gpt2 "github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type PromptBuilderHook int

const (
	PromptBuilderHookGlobalSystem PromptBuilderHook = iota
	PromptBuilderHookPreHistory
	PromptBuilderHookHistory
	PromptBuilderHookPostHistory
	PromptBuilderHookPreFocus
	PromptBuilderHookFocus
	PromptBuilderHookPostFocus
	PromptBuilderHookLast
)

type PromptBuilderHookFunc func(ctx context.Context, pb *PromptBuilder, req *openai.ChatCompletionRequest)

type PromptMessageSource func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*chat.Message], error)

func StaticMessageSource(items ...*chat.Message) PromptMessageSource {
	return func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*chat.Message], error) {
		return iterators.FromSlice(items), nil
	}
}

type PromptBuilderTool interface {
	ToolName() string
	ToolDefinition() *openai.FunctionDefinition
}

type simpleTool struct {
	definition *openai.FunctionDefinition
}

func WrapTool(def *openai.FunctionDefinition) PromptBuilderTool {
	return simpleTool{definition: def}
}

func (s simpleTool) ToolName() string                           { return s.definition.Name }
func (s simpleTool) ToolDefinition() *openai.FunctionDefinition { return s.definition }

type PromptBuilder struct {
	client       *openai.Client
	modelOptions gpt2.ModelOptions

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

		tokenizer: gpt2.GlobalModelTokenizer,
	}

	return b
}

func (b *PromptBuilder) WithClient(client *openai.Client) {
	b.client = client
}

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

func (b *PromptBuilder) WithModelOptions(opts gpt2.ModelOptions) {
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
		/*if b.tools[t.ToolName()] != nil {
			panic("duplicate tool: " + t.ToolName())
		}*/

		b.tools[t.ToolName()] = t
	}
}

func buildOrderedMap(m map[string]any) *orderedmap.OrderedMap {
	omap := orderedmap.New()

	for k, v := range m {
		omap.Set(k, v)
	}

	return omap
}

func (b *PromptBuilder) Build(ctx context.Context) openai.ChatCompletionRequest {
	var request openai.ChatCompletionRequest

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

	request.Functions = []openai.FunctionDefinition{
		/*{
			Name:        "CallNodeAction",
			Description: "Invokes a node action.",
			Parameters: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"path", "interface", "action", "arguments"},
				Properties: buildOrderedMap(map[string]any{
					"path": &jsonschema.Schema{
						Type:        "string",
						Description: "The path to the node to invoke the action on.",
					},
					"action": &jsonschema.Schema{
						Type:        "string",
						Description: "The name of the action to invoke.",
					},
					"arguments": &jsonschema.Schema{
						Type:        "object",
						Description: "The arguments to pass to the action.",
					},
				}),
			},
		},*/
		{
			Name:        "TraverseToNode",
			Description: "Traverse the given node.",
			Parameters: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"path"},
				Properties: buildOrderedMap(map[string]any{
					"path": &jsonschema.Schema{
						Type:        "string",
						Description: "The path of the node.",
					},
				}),
			},
		},
		{
			Name:        "InspectNode",
			Description: "Inspects the given node.",
			Parameters: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"path"},
				Properties: buildOrderedMap(map[string]any{
					"path": &jsonschema.Schema{
						Type:        "string",
						Description: "The path of the node.",
					},
				}),
			},
		},
		{
			Name:        "ShowAvailableFunctionsForNode",
			Description: "Show available functions and actions for a given node.",
			Parameters: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"path"},
				Properties: buildOrderedMap(map[string]any{
					"path": &jsonschema.Schema{
						Type:        "string",
						Description: "The path of the node.",
					},
				}),
			},
		},
		/*{
			Name:        "ListNodeEdges",
			Description: "Lists the edges of the given node.",
			Parameters: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"path"},
				Properties: buildOrderedMap(map[string]any{
					"path": &jsonschema.Schema{
						Type:        "string",
						Description: "The path of the node.",
					},
				}),
			},
		},*/
	}

	b.modelOptions.Apply(&request)

	if len(b.tools) > 0 {
		buffer := &bytes.Buffer{}
		_, _ = fmt.Fprintf(buffer, "Available Tools:\n")

		for _, tool := range b.tools {
			request.Functions = append(request.Functions, *tool.ToolDefinition())

			j, err := json.Marshal(tool.ToolDefinition().Parameters)

			if err != nil {
				panic(err)
			}

			_, _ = fmt.Fprintf(buffer, "- **%s:** %s `%s`\n", tool.ToolName(), tool.ToolDefinition().Description, string(j))
		}

		msg := chat.NewMessage(chat.MessageKindEmit)
		msg.From.Role = msn.RoleSystem
		msg.Text = buffer.String()

		b.AppendMessage(PromptBuilderHookPreFocus, msg)
	}

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

type ExecuteOptions struct {
	Client       *openai.Client
	ModelOptions gpt2.ModelOptions
}

func (o *ExecuteOptions) Apply(options ...ExecuteOption) {
	for _, opt := range options {
		opt(o)
	}
}

type ExecuteOption func(o *ExecuteOptions)

func (b *PromptBuilder) Execute(ctx context.Context, options ...ExecuteOption) (*PromptResponse, error) {
	var opts ExecuteOptions
	opts.Client = b.client
	opts.ModelOptions = b.modelOptions
	opts.Apply(options...)

	request := b.Build(ctx)

	trace := gpt2.CreateTrace(ctx, request)
	defer trace.End()

	err := func() error {
		ctx, span := tracer.Start(ctx, "openai.CreateChatCompletionStream")
		defer span.End()

		res, err := opts.Client.CreateChatCompletionStream(ctx, request)

		if err != nil {
			return err
		}

		defer res.Close()

		for {
			chunk, err := res.Recv()

			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				trace.ReportError(err)

				return err
			}

			trace.ConsumeOpenAI(chunk)
		}

		return nil
	}()

	if err != nil {
		return nil, err
	}

	return &PromptResponse{
		Raw: trace,

		Choices: lo.Map(trace.Choices, func(c openai.ChatCompletionChoice, _ int) PromptResponseChoice {
			msg := chat.NewMessage(chat.MessageKindEmit)
			msg.FromOpenAI(c.Message)

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
	}, nil
}

func (b *PromptBuilder) ExecuteAndParse(ctx context.Context, parser ResultParser, options ...ExecuteOption) error {
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
	stage := promptml.NewStage(root, gpt2.GlobalModelTokenizer)
	stage.MaxTokens = (gpt2.GlobalClient.MaxTokensForModel(o.Model) / 4 * 3) - int((float64(o.MaxTokens)*1.1)+10) - 256

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

func ExecuteWithClient(client *openai.Client) ExecuteOption {
	return func(o *ExecuteOptions) {
		o.Client = client
	}
}

func ExecuteWithModelOptions(opts gpt2.ModelOptions) ExecuteOption {
	return func(o *ExecuteOptions) {
		o.ModelOptions = o.ModelOptions.MergeWith(opts)
	}
}
