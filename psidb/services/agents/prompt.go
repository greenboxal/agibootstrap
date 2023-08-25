package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
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

type PromptMessageSource func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*Message], error)

func StaticMessageSource(items ...*Message) PromptMessageSource {
	return func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*Message], error) {
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
	request openai.ChatCompletionRequest

	tokenizer tokenizers.BasicTokenizer

	hooks    map[PromptBuilderHook][]PromptBuilderHookFunc
	messages map[PromptBuilderHook][]PromptMessageSource
	tools    map[string]PromptBuilderTool

	focus       *Message
	allMessages []*Message

	Context map[string]any
}

func NewPromptBuilder(base openai.ChatCompletionRequest) *PromptBuilder {
	b := &PromptBuilder{
		Context: map[string]any{},

		hooks:    map[PromptBuilderHook][]PromptBuilderHookFunc{},
		messages: map[PromptBuilderHook][]PromptMessageSource{},
		tools:    map[string]PromptBuilderTool{},

		tokenizer: gpt.GlobalModelTokenizer,
		request:   base,
	}

	return b
}

func (b *PromptBuilder) AllMessages() []*Message { return b.allMessages }

func (b *PromptBuilder) AddHook(hook PromptBuilderHook, fn PromptBuilderHookFunc) {
	b.hooks[hook] = append(b.hooks[hook], fn)
}

func (b *PromptBuilder) AppendMessageSources(hook PromptBuilderHook, srcs ...PromptMessageSource) {
	b.messages[hook] = append(b.messages[hook], srcs...)
}

func (b *PromptBuilder) AppendMessage(hook PromptBuilderHook, msg ...*Message) {
	b.AppendMessageSources(hook, StaticMessageSource(msg...))
}

func (b *PromptBuilder) AppendModelMessage(hook PromptBuilderHook, msg ...openai.ChatCompletionMessage) {
	mapped := lo.Map(msg, func(m openai.ChatCompletionMessage, _ int) *Message {
		msg := NewMessage(MessageKindEmit)
		msg.FromOpenAI(m)
		return msg
	})

	b.AppendMessageSources(hook, func(ctx context.Context, pb *PromptBuilder) (iterators.Iterator[*Message], error) {
		return iterators.FromSlice(mapped), nil
	})
}

func (b *PromptBuilder) SetFocus(msg *Message) { b.focus = msg }
func (b *PromptBuilder) GetFocus() *Message    { return b.focus }

func (b *PromptBuilder) WithModelOptions(opts ModelOptions) {
	opts.Apply(&b.request)
}

func (b *PromptBuilder) DisableTools() {
	b.request.FunctionCall = "none"
}

func (b *PromptBuilder) EnableTools() {
	b.request.FunctionCall = "auto"
}

func (b *PromptBuilder) ForceTool(name string) {
	b.request.FunctionCall = struct {
		Name string `json:"name"`
	}{Name: name}
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
	if len(b.tools) > 0 {
		b.request.Functions = []openai.FunctionDefinition{
			{
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

		buffer := &bytes.Buffer{}
		_, _ = fmt.Fprintf(buffer, "Available Tools:\n")

		for _, tool := range b.tools {
			j, err := json.Marshal(tool.ToolDefinition().Parameters)

			if err != nil {
				panic(err)
			}

			_, _ = fmt.Fprintf(buffer, "- **%s:** %s `%s`\n", tool.ToolName(), tool.ToolDefinition().Description, string(j))
		}

		msg := NewMessage(MessageKindEmit)
		msg.From.Role = msn.RoleSystem
		msg.Text = buffer.String()

		b.AppendMessage(PromptBuilderHookGlobalSystem, msg)
	}

	b.runHook(ctx, PromptBuilderHookGlobalSystem)
	b.runHook(ctx, PromptBuilderHookPreHistory)
	b.runHook(ctx, PromptBuilderHookHistory)
	b.runHook(ctx, PromptBuilderHookPostHistory)
	b.runHook(ctx, PromptBuilderHookPreFocus)
	b.runHook(ctx, PromptBuilderHookFocus)
	b.runHook(ctx, PromptBuilderHookPostFocus)
	b.runHook(ctx, PromptBuilderHookLast)

	return b.request
}

func (b *PromptBuilder) runHook(ctx context.Context, hook PromptBuilderHook) {
	for _, srcs := range b.messages[hook] {
		iter, err := srcs(ctx, b)

		if err != nil {
			panic(err)
		}

		for iter.Next() {
			msg := iter.Value()

			if hook != PromptBuilderHookFocus && msg == b.focus {
				continue
			}

			if msg.Kind == MessageKindEmit || msg.Kind == MessageKindError {
				b.allMessages = append(b.allMessages, msg)
				b.request.Messages = append(b.request.Messages, msg.ToOpenAI())
			}
		}
	}

	if hook == PromptBuilderHookFocus && b.focus != nil {
		b.allMessages = append(b.allMessages, b.focus)
		b.request.Messages = append(b.request.Messages, b.focus.ToOpenAI())
	}

	for _, fn := range b.hooks[hook] {
		fn(ctx, b, &b.request)
	}
}

func (b *PromptBuilder) Execute(ctx context.Context, client *openai.Client) (*PromptResponse, error) {
	request := b.Build(ctx)
	res, err := client.CreateChatCompletion(ctx, request)

	if err != nil {
		return nil, err
	}

	return &PromptResponse{
		Raw: &res,

		Choices: lo.Map(res.Choices, func(c openai.ChatCompletionChoice, _ int) PromptResponseChoice {
			msg := NewMessage(MessageKindEmit)
			msg.FromOpenAI(c.Message)

			choice := PromptResponseChoice{
				Index:   c.Index,
				Reason:  c.FinishReason,
				Message: msg,
			}

			if msg.FunctionCall != nil {
				switch msg.FunctionCall.Name {
				case "CallNodeAction":
					var args struct {
						Path      psi.Path        `json:"path"`
						ToolName  string          `json:"tool_name"`
						Arguments json.RawMessage `json:"arguments"`
					}

					if err := json.Unmarshal([]byte(msg.FunctionCall.Arguments), &args); err != nil {
						panic(err)
					}

					rawArgs, err := args.Arguments.MarshalJSON()

					if err != nil {
						panic(err)
					}

					choice.Tool = &PromptToolSelection{
						Focus:     args.Path,
						Name:      msg.FunctionCall.Name,
						Arguments: string(rawArgs),
						Tool:      b.tools[args.ToolName],
					}

				case "InspectNode":
					fallthrough
				case "ListNodeEdges":
					var args struct {
						Path psi.Path `json:"path"`
					}

					if err := json.Unmarshal([]byte(msg.FunctionCall.Arguments), &args); err != nil {
						panic(err)
					}

					choice.Tool = &PromptToolSelection{
						Focus: args.Path,
						Name:  msg.FunctionCall.Name,
					}

				default:
					choice.Tool = &PromptToolSelection{
						Name:      msg.FunctionCall.Name,
						Arguments: msg.FunctionCall.Arguments,
					}
				}
			}

			return choice
		}),
	}, nil
}

type PromptToolSelection struct {
	Focus     psi.Path `json:"focus"`
	Name      string   `json:"name"`
	Arguments string   `json:"arguments"`

	Tool PromptBuilderTool `json:"-"`
}

type PromptResponse struct {
	Raw     *openai.ChatCompletionResponse `json:"raw"`
	Choices []PromptResponseChoice         `json:"choices"`
}

type PromptResponseChoice struct {
	Index   int                         `json:"index"`
	Message *Message                    `json:"message"`
	Reason  openai.ChatCompletionReason `json:"reason"`

	Tool *PromptToolSelection `json:"tool"`
}
