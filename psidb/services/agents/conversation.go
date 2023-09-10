package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/invopop/jsonschema"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/samber/lo"
	openai2 "github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering/themes"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

var ConversationInterface = psi.DefineNodeInterface[IConversation]()
var ConversationType = psi.DefineNodeType[*Conversation](
	psi.WithInterfaceFromNode(chat.TopicSubscriberInterface),
	psi.WithInterfaceFromNode(chat.ChatInterface),
	psi.WithInterfaceFromNode(ConversationInterface),
)
var _ IConversation = (*Conversation)(nil)

// IConversation is an interface that defines the methods that a Conversation node must implement.
type IConversation interface {
	chat.IChat

	// OnMessageReceived is called when a message is received by the conversation.
	OnMessageReceived(ctx context.Context, req *OnMessageReceivedRequest) error
	// OnMessageSideEffect is called when a message is received by the conversation.
	OnMessageSideEffect(ctx context.Context, req *OnMessageSideEffectRequest) error
	// OnForkMerging is called when a fork is merged into the conversation.
	OnForkMerging(ctx context.Context, req *OnForkMergingRequest) error
}

type Conversation struct {
	psi.NodeBase

	Name string `json:"name"`

	BaseConversation *stdlib.Reference[*Conversation] `json:"base_conversation"`
	BaseMessage      *stdlib.Reference[*chat.Message] `json:"base_message"`
	BaseOptions      gpt.ModelOptions                 `json:"base_options"`

	IsMerged bool `json:"is_merged"`

	Client *openai.Client `json:"-" inject:""`
}

func (c *Conversation) PsiNodeName() string { return c.Name }

func (c *Conversation) HandleTopicMessage(ctx context.Context, message *stdlib.Reference[*chat.Message]) error {
	msg, err := message.Resolve(ctx)

	if err != nil {
		return err
	}

	if err := c.AcceptMessage(ctx, msg); err != nil {
		return err
	}

	if err := c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
		Message: message,
		Options: c.BuildDefaultOptions(),
	}); err != nil {
		return err
	}

	return c.Update(ctx)
}

func (c *Conversation) CreateThreadContext(ctx context.Context, message *chat.Message) *ThreadContext {
	return &ThreadContext{
		Ctx:          ctx,
		Client:       c.Client,
		ModelOptions: c.BuildDefaultOptions(),
		History:      c,
		Log:          c,
		BaseMessage:  message,
	}
}

func (c *Conversation) BuildDefaultOptions() gpt.ModelOptions {
	topP := float32(1.0)
	temperature := float32(0.0)
	model := "gpt-3.5-turbo-16k"
	maxTokens := 1024

	return gpt.ModelOptions{
		TopP:        &topP,
		Temperature: &temperature,
		Model:       &model,
		MaxTokens:   &maxTokens,
	}.MergeWith(c.BaseOptions)
}

func (c *Conversation) SliceMessages(ctx context.Context, from, to *chat.Message) ([]*chat.Message, error) {
	msgs, err := c.GetMessages(ctx, &chat.GetMessagesRequest{
		From: stdlib.Ref(from),
		To:   stdlib.Ref(to),
	})

	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (c *Conversation) MessageIterator(ctx context.Context) (iterators.Iterator[*chat.Message], error) {
	msgs, err := c.SliceMessages(ctx, nil, nil)

	if err != nil {
		return nil, err
	}

	return iterators.FromSlice(msgs), nil
}

func (c *Conversation) GetMessages(ctx context.Context, req *chat.GetMessagesRequest) ([]*chat.Message, error) {
	var err error
	var from, to *chat.Message
	var messages []*chat.Message

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

	if !req.SkipBaseHistory && c.BaseConversation != nil && (from == nil || from.Parent() != c) {
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

	ownMessages := psi.GetEdges(c, chat.ConversationMessageEdge)

	ownMessages = lo.Filter(ownMessages, func(item *chat.Message, index int) bool {
		if from != nil && item.Timestamp < from.Timestamp {
			return false
		}

		if to != nil && item.Timestamp > to.Timestamp {
			return false
		}

		return true
	})

	slices.SortFunc(ownMessages, func(i, j *chat.Message) bool {
		return strings.Compare(i.Timestamp, j.Timestamp) < 0
	})

	messages = append(messages, ownMessages...)

	return messages, nil
}

func (c *Conversation) SendMessage(ctx context.Context, req *chat.SendMessageRequest) (*chat.Message, error) {
	tx := coreapi.GetTransaction(ctx)

	m := chat.NewMessage(chat.MessageKindEmit)
	m.Timestamp = ""
	m.From = req.From
	m.Text = req.Text
	m.Attachments = req.Attachments
	m.Metadata = req.Metadata

	if req.Function != "" {
		m.FunctionCall = &chat.FunctionCall{
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

func (c *Conversation) ForkAsChatLog(ctx context.Context, baseMessage *chat.Message, options gpt.ModelOptions) (ChatLog, error) {
	fork, err := c.Fork(ctx, baseMessage, options)

	if err != nil {
		return nil, err
	}

	return fork, nil
}

func (c *Conversation) Fork(ctx context.Context, baseMessage *chat.Message, options gpt.ModelOptions) (*Conversation, error) {
	fork := &Conversation{}
	fork.BaseConversation = stdlib.Ref(c)
	fork.BaseMessage = stdlib.Ref(baseMessage)
	fork.Name = strconv.FormatInt(time.Now().UnixNano(), 10)
	fork.BaseOptions = c.BuildDefaultOptions().MergeWith(options)
	fork.Init(fork)
	fork.SetParent(c)

	if err := fork.Update(ctx); err != nil {
		return nil, err
	}

	joinMsg := chat.NewMessage(chat.MessageKindJoin)
	joinMsg.Attachments = []psi.Path{fork.CanonicalPath()}

	if _, err := c.addMessage(ctx, joinMsg); err != nil {
		return nil, err
	}

	return fork, nil
}

func (c *Conversation) Merge(ctx context.Context, focus *chat.Message) error {
	var mergeMsg *chat.Message

	baseMsg, err := c.BaseMessage.Resolve(ctx)

	if err != nil {
		return err
	}

	if !c.BaseConversation.IsEmpty() {
		base, err := c.BaseConversation.Resolve(ctx)

		if err != nil {
			return err
		}

		mergeMsg = chat.NewMessage(chat.MessageKindMerge)
		mergeMsg.Attachments = []psi.Path{focus.CanonicalPath(), base.CanonicalPath()}

		if _, err := c.addMessage(ctx, mergeMsg); err != nil {
			return err
		}

		c.IsMerged = true
		c.Invalidate()

		if err := coreapi.Dispatch(ctx, psi.Notification{
			Notifier:  c.CanonicalPath(),
			Notified:  base.CanonicalPath(),
			Interface: ConversationInterface.Name(),
			Action:    "OnForkMerging",
			Argument: OnForkMergingRequest{
				Fork:       stdlib.Ref(c),
				MergePoint: stdlib.Ref(mergeMsg),
			},
		}); err != nil {
			return err
		}
	}

	if len(baseMsg.ReplyTo) > 0 {
		msgs, err := c.SliceMessages(ctx, baseMsg, mergeMsg)

		if err != nil {
			return err
		}

		for _, msg := range msgs {
			if msg.Kind == chat.MessageKindEmit && msg.From.Role == msn.RoleAI && msg.FunctionCall == nil {
				for _, replyTo := range baseMsg.ReplyTo {
					if err := coreapi.Dispatch(ctx, psi.Notification{
						Notifier:  c.CanonicalPath(),
						Notified:  replyTo,
						Interface: chat.TopicSubscriberInterface.Name(),
						Action:    "HandleTopicMessage",
						Argument:  stdlib.Ref(msg),
					}); err != nil {
						return err
					}
				}
			}
		}
	}

	return c.Update(ctx)
}

type OnForkMergingRequest struct {
	Fork       *stdlib.Reference[*Conversation] `json:"fork" jsonschema:"title=Fork,description=The fork to merge"`
	MergePoint *stdlib.Reference[*chat.Message] `json:"merge_point" jsonschema:"title=Merge Point,description=The merge point"`
}

func (c *Conversation) OnForkMerging(ctx context.Context, req *OnForkMergingRequest) (err error) {
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
		if msg.Kind == chat.MessageKindError || (msg.Kind == chat.MessageKindEmit && msg.From.Role == msn.RoleAI && msg.FunctionCall == nil) {
			if _, err := c.addMessage(ctx, msg); err != nil {
				return err
			}
		}
	}

	return c.Update(ctx)
}

func (c *Conversation) OnMessageReceived(ctx context.Context, req *OnMessageReceivedRequest) (err error) {
	tx := coreapi.GetTransaction(ctx)
	opts := c.BuildDefaultOptions().MergeWith(req.Options)

	lastMessage, err := req.Message.Resolve(ctx)

	if err != nil {
		return err
	}

	fork := c

	if c.BaseConversation.IsEmpty() || lastMessage.From.Role == msn.RoleUser {
		fork, err = c.Fork(ctx, lastMessage, req.Options)

		if err != nil {
			return err
		}
	}

	handleError := func(cause error, dispatch bool) error {
		m := chat.NewMessage(chat.MessageKindError)
		m.From.ID = c.CanonicalPath()
		m.From.Name = "InspectNode"
		m.From.Role = msn.RoleFunction
		m.Text = fmt.Sprintf("Error: %s", cause)

		if req.ToolSelection != nil {
			m.From.Name = req.ToolSelection.Name
		}

		if _, err := fork.addMessage(ctx, m); err != nil {
			return err
		}

		if dispatch {
			if err := fork.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
				Message:       stdlib.Ref(m),
				Options:       req.Options,
				ToolSelection: req.ToolSelection,
			}); err != nil {
				return err
			}
		}

		return nil
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

	pb := NewPromptBuilder()
	pb.WithClient(c.Client)
	pb.WithModelOptions(opts)

	messages, err := fork.SliceMessages(ctx, nil, lastMessage)

	if err != nil {
		return err
	}

	pb.AppendMessage(PromptBuilderHookHistory, messages...)
	pb.SetFocus(lastMessage)

	pb.AppendModelMessage(PromptBuilderHookGlobalSystem, openai.ChatCompletionMessage{
		Role: "system",
		Content: fmt.Sprintf(`You are interfacing with a tree-structured database. The database contains nodes, and each node has a specific NodeType. Depending on its NodeType, a node can have different actions available to it.

QmYXZ is the root of the database. Any path like QmYXZ//foo/bar/baz will be resolved to QmYXZ//foo/bar/baz.
Depending on the NodeType of the selected node, you can perform specific actions on it.
You can read and write files by calling functions on the node. Consult the documentation for more information about the available functions and actions. Follow their declared JSONSchema.
Write messages for the user in Markdown.
Consult the documentation for more information about the available functions and actions. Follow their declared JSONSchema.
The user will send you prompts. These prompts might either be questions related to nodes or direct commands for you to carry out. Your goal is to understand the user's request and utilize the tools and actions at your disposal to satisfy their needs.
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
				Role: "system",
				Content: fmt.Sprintf(`You have attached a file. Here are the details:
File Path: %s
File Type: %s

%s`, node.CanonicalPath().String(), node.PsiNodeType().Name(), writer.String()),
			})
		}

		if req.ToolSelection != nil && req.ToolSelection.Name != "" {
			pb.ForceTool(req.ToolSelection.Name)
		} else if req.Options.ForceFunctionCall != nil {
			tool := *req.Options.ForceFunctionCall

			if tool != "" {
				pb.ForceTool(tool)
			}
		} else if lastMessage.From.Role == msn.RoleUser {
			if strings.HasPrefix(lastMessage.Text, "/enter-run-loop") {
				ffc := "ShowAvailableFunctionsForNode"
				pb.ForceTool(ffc)
			}
		}
	}

	result, err := pb.Execute(ctx)

	if err != nil {
		return handleError(err, true)
	}

	for _, choice := range result.Choices {
		if err := fork.consumeChoice(ctx, lastMessage, choice); err != nil {
			return err
		}
	}

	return c.Update(ctx)
}

func (c *Conversation) OnMessageSideEffect(ctx context.Context, req *OnMessageSideEffectRequest) (err error) {
	tx := coreapi.GetTransaction(ctx)

	handleError := func(cause error, dispatch bool) error {
		m := chat.NewMessage(chat.MessageKindError)
		m.From.ID = c.CanonicalPath()
		m.From.Name = "InspectNode"
		m.From.Role = msn.RoleFunction
		m.Text = fmt.Sprintf("Error: %s", cause)

		if req.ToolSelection != nil {
			m.From.Name = req.ToolSelection.Name
		}

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

		return nil
	}

	defer func() {
		if err != nil {
			if err := handleError(err, false); err != nil {
				panic(err)
			}
		}
	}()

	baseMessage, err := req.Message.Resolve(ctx)

	if err != nil {
		return err
	}

	switch req.ToolSelection.Name {
	case "CallNodeAction":
		var args struct {
			Path      psi.Path        `json:"path"`
			ToolName  string          `json:"tool_name"`
			Arguments json.RawMessage `json:"arguments"`
		}

		if err := json.Unmarshal([]byte(req.ToolSelection.Arguments), &args); err != nil {
			panic(err)
		}

		rawArgs, err := args.Arguments.MarshalJSON()

		if err != nil {
			panic(err)
		}

		req.ToolSelection.Arguments = string(rawArgs)
		req.ToolSelection.Name = args.ToolName

		if !args.Path.IsEmpty() {
			req.ToolSelection.Focus = stdlib.RefFromPath[psi.Node](args.Path)
		}

	case "ShowAvailableFunctionsForNode":
		fallthrough
	case "ActionForNode":
		var args struct {
			Path psi.Path `json:"path"`
		}

		if strings.TrimSpace(req.ToolSelection.Arguments) != "" {
			if err := json.Unmarshal([]byte(req.ToolSelection.Arguments), &args); err != nil {
				panic(err)
			}
		}

		if args.Path.IsEmpty() && len(baseMessage.Attachments) > 0 {
			args.Path = baseMessage.Attachments[0]
		}

		target, err := tx.Resolve(ctx, args.Path)

		if err != nil {
			return handleError(err, true)
		}

		return c.showAvailableActions(ctx, req, target)

	case "TraverseToNode":
		fallthrough
	case "TraverseTo":
		fallthrough
	case "InspectNode":
		var args struct {
			Path psi.Path `json:"path"`
		}

		if err := json.Unmarshal([]byte(req.ToolSelection.Arguments), &args); err != nil {
			panic(err)
		}

		if args.Path.IsEmpty() && len(baseMessage.Attachments) > 0 {
			args.Path = baseMessage.Attachments[0]
		}

		target, err := tx.Resolve(ctx, args.Path)

		if err != nil {
			return handleError(err, true)
		}

		return c.inspectNode(ctx, req, target)
	}

	target, err := req.ToolSelection.Focus.Resolve(ctx)

	if err != nil {
		return handleError(err, true)
	}

	ifaceName, actionName, _ := strings.Cut(req.ToolSelection.Name, "_QZQZ_")

	iface := target.PsiNodeType().Interface(ifaceName)

	if iface == nil {
		return handleError(fmt.Errorf("interface %s not found", ifaceName), true)
	}

	action := iface.Action(actionName)

	if action == nil {
		return handleError(fmt.Errorf("action %s not found", actionName), true)
	}

	if action.RequestType() != nil {
		form := NewForm(action.RequestType().JsonSchema())

		if err := form.ParseJson([]byte(req.ToolSelection.Arguments)); err != nil {
			return handleError(err, true)
		}

		/*valid, err := form.Validate()

		if err != nil {
			return handleError(err, true)
		}

		if !valid || req.ToolSelection.Arguments == "" {
			tctx := c.CreateThreadContext(ctx, baseMessage)
			tctx, err = tctx.Fork(ctx, baseMessage, req.Options)

			if err != nil {
				return err
			}

			if req.ToolSelection.Arguments == "" {
				if err := form.FillAll(tctx); err != nil {
					return handleError(err, true)
				}
			} else {
				if ok, err := form.Fix(tctx); err != nil {
					return handleError(err, true)
				} else if !ok {
					return c.dispatchSideEffect(ctx, c.CanonicalPath(), *req)
				}
			}
		}*/

		fixed, err := form.ToJSON()

		if err != nil {
			return err
		}

		req.ToolSelection.Arguments = string(fixed)
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
		} else if result != nil {
			if err := ipld.EncodeStreaming(writer, typesystem.Wrap(result), dagjson.Encode); err != nil {
				writer.Write([]byte(fmt.Sprintf("Error: %v", err)))
				return
			}
		}

		writer.Write([]byte("\n\nContinue."))
	}()

	replyMessage := chat.NewMessage(chat.MessageKindEmit)
	replyMessage.From.ID = c.CanonicalPath()
	replyMessage.From.Name = req.ToolSelection.Name
	replyMessage.From.Role = msn.RoleFunction
	replyMessage.Text = writer.String()
	replyMessage.Attachments = attachments

	replyMessage, err = c.addMessage(ctx, replyMessage)

	if err != nil {
		return err
	}

	return c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
		Message: stdlib.Ref(replyMessage),
		Options: req.Options,
	})
}

func (c *Conversation) showAvailableActions(ctx context.Context, req *OnMessageSideEffectRequest, target psi.Node) error {
	writer := &bytes.Buffer{}
	actionMap := map[string]*jsonschema.Schema{}

	for _, iface := range target.PsiNodeType().Interfaces() {
		for _, action := range iface.Interface().Actions() {
			var schema *jsonschema.Schema

			if action.RequestType != nil {
				schema = action.RequestType.JsonSchema()
			}

			if schema == nil {
				schema = &jsonschema.Schema{Type: "object"}
			}

			actionName := fmt.Sprintf("%s_QZQZ_%s", iface.Interface().Name(), action.Name)
			actionMap[actionName] = schema
		}
	}

	_, _ = fmt.Fprintf(writer, "**Path:** %s\n", target.CanonicalPath().String())
	_, _ = fmt.Fprintf(writer, "**Node Type:** %s\n", target.PsiNodeType().Name())
	_, _ = fmt.Fprintf(writer, "# Actions\n\n```json\n")

	if err := json.NewEncoder(writer).Encode(actionMap); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(writer, "\n```\n")

	replyMessage := chat.NewMessage(chat.MessageKindEmit)
	replyMessage.From.ID = c.CanonicalPath()
	replyMessage.From.Name = req.ToolSelection.Name
	replyMessage.From.Role = msn.RoleFunction
	replyMessage.Text = writer.String()
	replyMessage.Attachments = []psi.Path{target.CanonicalPath()}

	replyMessage, err := c.addMessage(ctx, replyMessage)

	if err != nil {
		return err
	}

	return c.dispatchModel(ctx, c.CanonicalPath(), OnMessageReceivedRequest{
		Message: stdlib.Ref(replyMessage),
		Options: req.Options,
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

	replyMessage := chat.NewMessage(chat.MessageKindEmit)
	replyMessage.From.ID = c.CanonicalPath()
	replyMessage.From.Name = req.ToolSelection.Name
	replyMessage.From.Role = msn.RoleFunction
	replyMessage.Text = writer.String()
	replyMessage.Attachments = []psi.Path{target.CanonicalPath()}

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

func (c *Conversation) AcceptChoice(ctx context.Context, baseMessage *chat.Message, choice PromptResponseChoice) error {
	return c.consumeChoice(ctx, baseMessage, choice)
}

func (c *Conversation) AcceptMessage(ctx context.Context, msg *chat.Message) error {
	_, err := c.addMessage(ctx, msg)

	if err != nil {
		return err
	}

	return nil
}

var logger = logging.GetLogger("psi/mod/chat")

func (c *Conversation) addMessage(ctx context.Context, m *chat.Message) (*chat.Message, error) {
	if m.Timestamp == "" {
		m.Timestamp = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	if err := m.Update(ctx); err != nil {
		return nil, err
	}

	if m.Parent() == nil {
		if m.Kind == chat.MessageKindError {
			logger.Error(m.Text)
		}

		m.SetParent(c)
	}

	c.SetEdge(chat.ConversationMessageEdge.Named(m.Timestamp), m)

	return m, c.Update(ctx)
}

func (c *Conversation) consumeChoice(ctx context.Context, baseMessage *chat.Message, choice PromptResponseChoice) error {
	m, err := c.addMessage(ctx, choice.Message)

	if err != nil {
		return err
	}

	if choice.Reason == openai2.FinishReasonFunctionCall || choice.Tool != nil && choice.Tool.Name != "" {
		if (choice.Tool.Focus == nil || choice.Tool.Focus.IsEmpty()) && len(baseMessage.Attachments) > 0 {
			if choice.Tool == nil {
				choice.Tool = &PromptToolSelection{}
			}

			choice.Tool.Focus = stdlib.RefFromPath[psi.Node](baseMessage.Attachments[0])
		}

		if err := c.dispatchSideEffect(ctx, c.CanonicalPath(), OnMessageSideEffectRequest{
			Message:       stdlib.Ref(m),
			Options:       c.BuildDefaultOptions(),
			ToolSelection: choice.Tool,
		}); err != nil {
			return err
		}
	} else if c.BaseConversation != nil {
		if err := c.Merge(ctx, m); err != nil {
			return err
		}
	}

	return nil
}

type OnMessageReceivedRequest struct {
	Message       *stdlib.Reference[*chat.Message] `json:"message" jsonschema:"title=Message,description=The message received"`
	Options       gpt.ModelOptions                 `json:"options" jsonschema:"title=Model Options,description=The options for the model"`
	ToolSelection *PromptToolSelection             `json:"tool_selection" jsonschema:"title=Tool Selection,description=The tool selection,optional"`
}

type OnMessageSideEffectRequest struct {
	Message *stdlib.Reference[*chat.Message] `json:"message" jsonschema:"title=Message,description=The message received"`
	Options gpt.ModelOptions                 `json:"options" jsonschema:"title=Model Options,description=The options for the model"`

	ToolSelection *PromptToolSelection `json:"tool_selection" jsonschema:"title=Tool Selection,description=The tool selection"`
}
