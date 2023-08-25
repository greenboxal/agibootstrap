package agents

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/samber/lo"
	openai2 "github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
	"github.com/greenboxal/agibootstrap/pkg/typesystem"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
)

var ConversationInterface = psi.DefineNodeInterface[IConversation]()
var ConversationType = psi.DefineNodeType[*Conversation](psi.WithInterfaceFromNode(ConversationInterface))
var ConversationMessageEdge = psi.DefineEdgeType[*Message]("message")
var _ IConversation = (*Conversation)(nil)

type IConversation interface {
	GetMessages(ctx context.Context, req *GetMessagesRequest) ([]*Message, error)
	SendMessage(ctx context.Context, req *SendMessageRequest) (*Message, error)

	OnMessageReceived(ctx context.Context, req *OnMessageReceivedRequest) error
	OnMessageSideEffect(ctx context.Context, req *OnMessageSideEffectRequest) error
	OnForkMerging(ctx context.Context, req *OnForMergingRequest) error
}

type Conversation struct {
	psi.NodeBase

	Name string `json:"name"`

	BaseConversation *stdlib.Reference[*Conversation] `json:"base_conversation"`
	BaseMessage      *stdlib.Reference[*Message]      `json:"base_message"`

	IsMerged bool `json:"is_merged"`

	Client *openai.Client `json:"-" inject:""`
}

func (c *Conversation) PsiNodeName() string { return c.Name }

type GetMessagesRequest struct {
	From *stdlib.Reference[*Message] `json:"from"`
	To   *stdlib.Reference[*Message] `json:"to"`
}

func (c *Conversation) GetMessages(ctx context.Context, req *GetMessagesRequest) ([]*Message, error) {
	var err error
	var from, to *Message

	if !req.From.IsEmpty() {
		from, err = req.From.Resolve(ctx)

		if err != nil {
			return nil, err
		}
	}

	if !req.To.IsEmpty() {
		to, err = req.To.Resolve(ctx)

		if err != nil {
			return nil, err
		}
	}

	return c.SliceMessages(ctx, from, to)
}

func (c *Conversation) SliceMessages(ctx context.Context, from, to *Message) ([]*Message, error) {
	var messages []*Message

	if c.BaseConversation != nil && (from == nil || from.Parent() != c) {
		baseConversation, err := c.BaseConversation.Resolve(ctx)

		if err != nil {
			return nil, err
		}

		baseMessage, err := c.BaseMessage.Resolve(ctx)

		if err != nil {
			return nil, err
		}

		baseMessages, err := baseConversation.SliceMessages(ctx, from, baseMessage)

		if err != nil {
			return nil, err
		}

		messages = append(messages, baseMessages...)
	}

	ownMessages := psi.GetEdges(c, ConversationMessageEdge)

	slices.SortFunc(ownMessages, func(i, j *Message) bool {
		return strings.Compare(i.Timestamp, j.Timestamp) < 0
	})

	if from != nil || to != nil {
		for i, msg := range ownMessages {
			if msg == from {
				ownMessages = ownMessages[i:]

				if to == nil {
					break
				}
			} else if msg == to {
				ownMessages = ownMessages[:i]
				break
			}
		}
	}

	messages = append(messages, ownMessages...)

	return messages, nil
}

func (c *Conversation) SendMessage(ctx context.Context, req *SendMessageRequest) (*Message, error) {
	tx := coreapi.GetTransaction(ctx)

	m := NewMessage(MessageKindEmit)
	m.Timestamp = ""
	m.From = req.From
	m.Text = req.Text
	m.Attachments = req.Attachments
	m.Metadata = req.Metadata

	if req.Function != "" {
		m.FunctionCall = &FunctionCall{
			Name:      req.Function,
			Arguments: req.FunctionArguments,
		}
	}

	m, err := c.addMessage(ctx, m)

	if err != nil {
		return nil, err
	}

	if err := tx.Notify(ctx, psi.Notification{
		Notifier:  c.CanonicalPath(),
		Notified:  c.CanonicalPath(),
		Interface: ConversationInterface.Name(),
		Action:    "OnMessageReceived",
		Argument: &OnMessageReceivedRequest{
			Message: stdlib.Ref(m),
			Options: req.ModelOptions,
		},
	}); err != nil {
		return nil, err
	}

	return m, c.Update(ctx)
}

func (c *Conversation) addMessage(ctx context.Context, m *Message) (*Message, error) {
	if m.Timestamp == "" {
		m.Timestamp = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	if err := m.Update(ctx); err != nil {
		return nil, err
	}

	if m.Parent() == nil {
		m.SetParent(c)
	}

	c.SetEdge(ConversationMessageEdge.Named(m.Timestamp), m)

	return m, c.Update(ctx)
}

func (c *Conversation) Fork(ctx context.Context, baseMessage *Message) (*Conversation, error) {
	fork := &Conversation{}
	fork.BaseConversation = stdlib.Ref(c)
	fork.BaseMessage = stdlib.Ref(baseMessage)
	fork.Name = strconv.FormatInt(time.Now().UnixNano(), 10)
	fork.Init(fork)
	fork.SetParent(c)

	if err := fork.Update(ctx); err != nil {
		return nil, err
	}

	joinMsg := NewMessage(MessageKindJoin)
	joinMsg.Attachments = []psi.Path{fork.CanonicalPath()}

	if _, err := c.addMessage(ctx, joinMsg); err != nil {
		return nil, err
	}

	return fork, nil
}

func (c *Conversation) Merge(ctx context.Context, focus *Message) error {
	if c.BaseConversation.IsEmpty() {
		return nil
	}

	base, err := c.BaseConversation.Resolve(ctx)

	if err != nil {
		return err
	}

	mergeMsg := NewMessage(MessageKindMerge)
	mergeMsg.Attachments = []psi.Path{focus.CanonicalPath(), base.CanonicalPath()}

	if _, err := c.addMessage(ctx, mergeMsg); err != nil {
		return err
	}

	c.IsMerged = true

	if err := c.Update(ctx); err != nil {
		return err
	}

	return coreapi.Dispatch(ctx, psi.Notification{
		Notifier:  c.CanonicalPath(),
		Notified:  base.CanonicalPath(),
		Interface: ConversationInterface.Name(),
		Action:    "OnForkMerging",
		Argument: OnForMergingRequest{
			Fork:       stdlib.Ref(c),
			MergePoint: stdlib.Ref(mergeMsg),
		},
	})
}

type OnForMergingRequest struct {
	Fork       *stdlib.Reference[*Conversation] `json:"fork" jsonschema:"title=Fork,description=The fork to merge"`
	MergePoint *stdlib.Reference[*Message]      `json:"merge_point" jsonschema:"title=Merge Point,description=The merge point"`
}

func (c *Conversation) OnForkMerging(ctx context.Context, req *OnForMergingRequest) (err error) {
	fork, err := req.Fork.Resolve(ctx)

	if err != nil {
		return err
	}

	baseMessage, err := fork.BaseMessage.Resolve(ctx)

	if err != nil {
		return err
	}

	mergePoint, err := req.MergePoint.Resolve(ctx)

	if err != nil {
		return err
	}

	msgs, err := fork.SliceMessages(ctx, baseMessage, mergePoint)

	if err != nil {
		return err
	}

	for _, msg := range msgs {
		if msg == baseMessage || msg == mergePoint {
			continue
		}

		if msg.From.Role == msn.RoleSystem {
			continue
		}

		if _, err := c.addMessage(ctx, msg); err != nil {
			return err
		}
	}

	return c.Update(ctx)
}

func (c *Conversation) OnMessageReceived(ctx context.Context, req *OnMessageReceivedRequest) (err error) {
	tx := coreapi.GetTransaction(ctx)

	handleError := func(cause error, dispatch bool) error {
		m := NewMessage(MessageKindError)
		m.From.ID = "function"
		m.From.Name = req.ToolSelection.Name
		m.From.Role = msn.RoleFunction
		m.Text = fmt.Sprintf("Error: %v", cause)

		if _, err := c.addMessage(ctx, m); err != nil {
			return err
		}

		if dispatch {
			if err := c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
				Message:       stdlib.Ref(m),
				Options:       req.Options,
				ToolSelection: req.ToolSelection,
			}); err != nil {
				return err
			}
		}

		return cause
	}

	defer func() {
		if err != nil {
			if err := handleError(err, false); err != nil {
				panic(err)
			}
		}

		if err := c.Update(ctx); err != nil {
			panic(err)
		}
	}()

	lastMessage, err := req.Message.Resolve(ctx)

	if err != nil {
		return err
	}

	fork := c

	if c.BaseConversation.IsEmpty() || lastMessage.From.Role == msn.RoleUser {
		fork, err = c.Fork(ctx, lastMessage)

		if err != nil {
			return err
		}
	}

	pb := NewPromptBuilder(openai.ChatCompletionRequest{
		TopP:        1.0,
		Temperature: 1.0,
		Model:       "gpt-3.5-turbo-16k",
		MaxTokens:   4096,
		N:           1,
	})

	pb.WithModelOptions(req.Options)

	pb.AddHook(PromptBuilderHookLast, func(ctx context.Context, pb *PromptBuilder, req *openai.ChatCompletionRequest) {
		focusBase, err := strconv.ParseInt(lastMessage.Timestamp, 10, 64)

		if err != nil {
			focusBase = time.Now().UnixNano()
		}

		focusBase -= 100000

		for i, msg := range pb.AllMessages() {
			if msg.Parent() == nil {
				if msg.Timestamp == "" {
					msg.Timestamp = strconv.FormatInt(focusBase+int64(i), 10)
				}

				if _, err := fork.addMessage(ctx, msg); err != nil {
					panic(err)
				}
			}
		}
	})

	messages, err := fork.SliceMessages(ctx, nil, lastMessage)

	if err != nil {
		return err
	}

	messages = lo.Filter(messages, func(msg *Message, _ int) bool {
		return msg.From.Role != msn.RoleSystem || (msg.Parent() == nil || msg.Parent() == fork)
	})

	pb.AppendMessage(PromptBuilderHookHistory, messages...)
	pb.SetFocus(lastMessage)

	pb.AppendModelMessage(PromptBuilderHookGlobalSystem, openai.ChatCompletionMessage{
		Role: "system",
		Content: fmt.Sprintf(`You are interfacing with a tree-structured database. The database contains nodes, and each node has a specific NodeType. Depending on its NodeType, a node can have different actions available to it. You, the agent, can:

Browse the Tree: Use commands or function calls to traverse the tree database.
Select a Node: You can pick a specific node to operate on.
Execute Actions: Depending on the NodeType of the selected node, you can perform specific actions on it. Consult the documentation for more information.
The user will send you prompts. These prompts might either be questions related to nodes or direct commands for you to carry out. Your goal is to understand the user's request and utilize the tools and actions at your disposal to satisfy their needs.
Write messages for the user preferably in Markdown.
`),
	})

	if len(lastMessage.Attachments) > 0 {
		pb.AppendModelMessage(PromptBuilderHookPreFocus, openai.ChatCompletionMessage{
			Role: "system",
			Content: fmt.Sprintf(`
I noticed that your message contains attachments. To manipulate them, please refer to the available functions in the documentation.
`),
		})

		for _, attachment := range lastMessage.Attachments {
			node, err := tx.Resolve(ctx, attachment)

			if err != nil {
				return err
			}

			actions := buildActionsFor(node)

			for _, action := range actions {
				def := openai.FunctionDefinition{
					Name:        strings.Replace(action.Name, ".", "_", -1),
					Description: action.Description,
					Parameters:  action.Parameters,
				}

				pb.WithTools(WrapTool(&def))
			}

			writer := &bytes.Buffer{}

			if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", node); err != nil {
				return err
			}

			pb.AppendModelMessage(PromptBuilderHookPreFocus, openai.ChatCompletionMessage{
				Role: "user",
				Content: fmt.Sprintf(`You have attached a file. Here are the details:
File Path: %s
File Type: %s

%s`, node.CanonicalPath().String(), node.PsiNodeType().Name(), writer.String()),
			})
		}

		if req.ToolSelection != nil && req.ToolSelection.Name != "" {
			pb.ForceTool(req.ToolSelection.Name)
		} else {
			pb.ForceTool("CallNodeAction")
		}
	}

	result, err := pb.Execute(ctx, c.Client)

	if err != nil {
		return handleError(err, true)
	}

	for _, choice := range result.Choices {
		m, err := fork.addMessage(ctx, choice.Message)

		if err != nil {
			return nil
		}

		if choice.Reason == openai2.FinishReasonFunctionCall && choice.Tool != nil {
			if choice.Tool.Focus.IsEmpty() && len(lastMessage.Attachments) > 0 {
				choice.Tool.Focus = lastMessage.Attachments[0]
			}

			if err := fork.dispatchSideEffect(ctx, c.CanonicalPath(), OnMessageSideEffectRequest{
				Message:       stdlib.Ref(m),
				Options:       req.Options,
				ToolSelection: choice.Tool,
			}); err != nil {
				return err
			}
		} else if fork.BaseConversation != nil {
			if err := fork.Merge(ctx, m); err != nil {
				return err
			}
		}
	}

	return c.Update(ctx)
}

func (c *Conversation) OnMessageSideEffect(ctx context.Context, req *OnMessageSideEffectRequest) (err error) {
	tx := coreapi.GetTransaction(ctx)

	handleError := func(cause error, dispatch bool) error {
		m := NewMessage(MessageKindError)
		m.From.ID = "function"
		m.From.Name = req.ToolSelection.Name
		m.From.Role = msn.RoleFunction
		m.Text = fmt.Sprintf("Error: %v", cause)

		if _, err := c.addMessage(ctx, m); err != nil {
			return err
		}

		if dispatch {
			if err := c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
				Message:       stdlib.Ref(m),
				Options:       req.Options,
				ToolSelection: req.ToolSelection,
			}); err != nil {
				return err
			}
		}

		return cause
	}

	if err := c.Update(ctx); err != nil {
		panic(err)
	}

	defer func() {
		if err != nil {
			if err := handleError(err, false); err != nil {
				panic(err)
			}
		}
	}()

	target, err := tx.Resolve(ctx, req.ToolSelection.Focus)

	if err != nil {
		return handleError(err, true)
	}

	switch req.ToolSelection.Name {
	case "InspectNode":
		return c.inspectNode(ctx, req, target)
	}

	ifaceName, actionName, _ := strings.Cut(req.ToolSelection.Name, "___")

	if req.ToolSelection.Arguments == "" {
		iface := target.PsiNodeType().Interface(ifaceName)

		if iface == nil {
			return handleError(fmt.Errorf("interface %s not found", ifaceName), true)
		}

		action := iface.Action(actionName)

		if action == nil {
			return handleError(fmt.Errorf("action %s not found", actionName), true)
		}

		if action.RequestType() != nil {
			actionSchema, err := action.RequestType().JsonSchema().MarshalJSON()

			if err != nil {
				return err
			}

			queryMsg := NewMessage(MessageKindEmit)
			queryMsg.From.ID = "function"
			queryMsg.From.Name = req.ToolSelection.Name
			queryMsg.From.Role = msn.RoleFunction
			queryMsg.Text = fmt.Sprintf("```\n%s\n```\n\nFill out the parameters according to the schema.", string(actionSchema))

			queryMsg, err = c.addMessage(ctx, queryMsg)

			if err != nil {
				return err
			}

			return c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
				Message:       stdlib.Ref(queryMsg),
				Options:       req.Options,
				ToolSelection: req.ToolSelection,
			})
		}
	}

	not := psi.Notification{
		Notifier:  c.CanonicalPath(),
		Notified:  target.CanonicalPath(),
		Interface: ifaceName,
		Action:    actionName,
		Params:    []byte(req.ToolSelection.Arguments),
	}

	writer := &bytes.Buffer{}
	attachments := []psi.Path{target.CanonicalPath()}

	func() {
		defer func() {
			if err := recover(); err != nil {
				writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
			}
		}()

		result, err := not.Apply(ctx, target)

		if err != nil {
			writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
			return
		}

		if node, ok := result.(psi.Node); ok {
			if node.Parent() == nil {
				node.SetParent(target)

				if err := node.Update(ctx); err != nil {
					writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
					return
				}
			}

			attachments = append(attachments, node.CanonicalPath())

			if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", node); err != nil {
				writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
				return
			}
		} else if it, ok := result.(psi.EdgeIterator); ok {
			for it.Next() {
				edge := it.Value()
				node := edge.To()

				attachments = append(attachments, node.CanonicalPath())

				if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", node); err != nil {
					writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
					return
				}
			}
		} else if it, ok := result.(psi.NodeIterator); ok {
			for it.Next() {
				node := it.Value()

				attachments = append(attachments, node.CanonicalPath())

				if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", node); err != nil {
					writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
					return
				}
			}
		} else {
			if err := ipld.EncodeStreaming(writer, typesystem.Wrap(result), dagjson.Encode); err != nil {
				writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
				return
			}
		}
	}()

	replyMessage := NewMessage(MessageKindEmit)
	replyMessage.From.ID = "function"
	replyMessage.From.Name = req.ToolSelection.Name
	replyMessage.From.Role = msn.RoleFunction
	replyMessage.Text = writer.String()
	replyMessage.Attachments = attachments

	replyMessage, err = c.addMessage(ctx, replyMessage)

	if err != nil {
		return err
	}

	return c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
		Message:       stdlib.Ref(replyMessage),
		Options:       req.Options,
		ToolSelection: req.ToolSelection,
	})
}

func (c *Conversation) inspectNode(ctx context.Context, req *OnMessageSideEffectRequest, target psi.Node) error {
	tx := coreapi.GetTransaction(ctx)
	writer := &bytes.Buffer{}

	_, _ = fmt.Fprintf(writer, "**Path:** %s\n", target.CanonicalPath().String())
	_, _ = fmt.Fprintf(writer, "**Node Type:** %s\n", target.PsiNodeType().Name())

	_, _ = fmt.Fprintf(writer, "# Edges\n\n")
	for edges := target.Edges(); edges.Next(); {
		edge := edges.Value()
		to, err := edge.ResolveTo(ctx)

		if err != nil {
			writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
		} else {
			_, _ = fmt.Fprintf(writer, "- **%s:** %s\n", edge.Key(), to.PsiNodeType().Name())
		}
	}

	_, _ = fmt.Fprintf(writer, "# Node\n\n")
	if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", target); err != nil {
		writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
	}

	replyMessage := NewMessage(MessageKindEmit)
	replyMessage.From.ID = "function"
	replyMessage.From.Name = req.ToolSelection.Name
	replyMessage.From.Role = msn.RoleFunction
	replyMessage.Text = writer.String()

	replyMessage, err := c.addMessage(ctx, replyMessage)

	if err != nil {
		return err
	}

	return tx.Notify(ctx, psi.Notification{
		Notifier:  c.CanonicalPath(),
		Notified:  c.CanonicalPath(),
		Interface: ConversationInterface.Name(),
		Action:    "OnMessageReceived",
		Argument: &OnMessageReceivedRequest{
			Message: stdlib.Ref(replyMessage),
			Options: req.Options,
		},
	})
}

func (c *Conversation) dispatchModel(ctx context.Context, requestor psi.Path, request OnMessageReceivedRequest) error {
	tx := coreapi.GetTransaction(ctx)

	if err := tx.Notify(ctx, psi.Notification{
		Notifier:  requestor,
		Notified:  c.CanonicalPath(),
		Interface: ConversationInterface.Name(),
		Action:    "OnMessageReceived",
		Argument:  request,
	}); err != nil {
		return err
	}

	return nil
}

func (c *Conversation) dispatchSideEffect(ctx context.Context, path psi.Path, request OnMessageSideEffectRequest) error {
	tx := coreapi.GetTransaction(ctx)

	if err := tx.Notify(ctx, psi.Notification{
		Notifier:  path,
		Notified:  c.CanonicalPath(),
		Interface: ConversationInterface.Name(),
		Action:    "OnMessageSideEffect",
		Argument:  request,
	}); err != nil {
		return err
	}

	return nil
}

type SendMessageRequest struct {
	Timestamp   string         `json:"timestamp" jsonschema:"title=Timestamp,description=The timestamp of the message"`
	From        UserHandle     `json:"from" jsonschema:"title=From,description=The user who sent the message"`
	Text        string         `json:"text" jsonschema:"title=Text,description=The text content of the message"`
	Attachments []psi.Path     `json:"attachments" jsonschema:"title=Attachments,description=The attachments included in the message"`
	Metadata    map[string]any `json:"metadata" jsonschema:"title=Metadata,description=The metadata associated with the message"`

	ModelOptions ModelOptions `json:"model_options" jsonschema:"title=Model Options,description=The options for the model"`

	Function          string `json:"function" jsonschema:"title=Function,description=The function to execute"`
	FunctionArguments string `json:"function_arguments" jsonschema:"title=Function Arguments,description=The arguments for the function"`
}

type ModelOptions struct {
	Model            *string        `json:"model" jsonschema:"title=Model,description=The model used for the conversation"`
	MaxTokens        *int           `json:"max_tokens" jsonschema:"title=Max Tokens,description=The maximum number of tokens for the model"`
	Temperature      *float32       `json:"temperature" jsonschema:"title=Temperature,description=The temperature setting for the model"`
	TopP             *float32       `json:"top_p" jsonschema:"title=Top P,description=The top P setting for the model"`
	FrequencyPenalty *float32       `json:"frequency_penalty" jsonschema:"title=Frequency Penalty,description=The frequency penalty setting for the model"`
	PresencePenalty  *float32       `json:"presence_penalty" jsonschema:"title=Presence Penalty,description=The presence penalty setting for the model"`
	Stop             []string       `json:"stop" jsonschema:"title=Stop,description=The stop words for the model"`
	LogitBias        map[string]int `json:"logit_bias" jsonschema:"title=Logit Bias,description=The logit bias setting for the model"`
}

func (o ModelOptions) Apply(req *openai.ChatCompletionRequest) {
	if o.Model != nil {
		req.Model = *o.Model
	}

	if o.MaxTokens != nil {
		req.MaxTokens = *o.MaxTokens
	}

	if o.Temperature != nil {
		req.Temperature = *o.Temperature
	}

	if o.TopP != nil {
		req.TopP = *o.TopP
	}

	if o.FrequencyPenalty != nil {
		req.FrequencyPenalty = *o.FrequencyPenalty
	}

	if o.PresencePenalty != nil {
		req.PresencePenalty = *o.PresencePenalty
	}

	if o.Stop != nil {
		req.Stop = o.Stop
	}

	if o.LogitBias != nil {
		req.LogitBias = o.LogitBias
	}
}

type OnMessageReceivedRequest struct {
	Message       *stdlib.Reference[*Message] `json:"message" jsonschema:"title=Message,description=The message received"`
	Options       ModelOptions                `json:"options" jsonschema:"title=Model Options,description=The options for the model"`
	ToolSelection *PromptToolSelection        `json:"tool_selection" jsonschema:"title=Tool Selection,description=The tool selection,optional"`
}

type OnMessageSideEffectRequest struct {
	Message *stdlib.Reference[*Message] `json:"message" jsonschema:"title=Message,description=The message received"`
	Options ModelOptions                `json:"options" jsonschema:"title=Model Options,description=The options for the model"`

	ToolSelection *PromptToolSelection `json:"tool_selection" jsonschema:"title=Tool Selection,description=The tool selection"`
}
