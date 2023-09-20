package agents

import (
	"context"
	"encoding/json"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
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
	// UpdateTitle updates the title of the conversation.
	UpdateTitle(ctx context.Context, req *UpdateTitleRequest) error
}

type Conversation struct {
	psi.NodeBase

	Name string `json:"name"`

	Title            string `json:"title"`
	IsTitleTemporary bool   `json:"is_title_temporary"`

	BaseConversation *stdlib.Reference[*Conversation] `json:"base_conversation"`
	BaseMessage      *stdlib.Reference[*chat.Message] `json:"base_message"`
	BaseOptions      gpt.ModelOptions                 `json:"base_options"`

	IsMerged bool `json:"is_merged"`

	TraceTags []string `json:"trace_tags,omitempty"`

	Client *openai.Client `json:"-" inject:""`
}

func (c *Conversation) PsiNodeName() string { return c.Name }

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

type FunctionCallArgumentHolder struct {
	Choices []json.RawMessage
}

func (h *FunctionCallArgumentHolder) Prepare(i int) []any {
	h.Choices = make([]json.RawMessage, i)

	refs := make([]any, i)

	for i := range refs {
		refs[i] = &h.Choices[i]
	}

	return refs
}
