package agents

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	openai2 "github.com/sashabaranov/go-openai"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
)

var ConversationInterface = psi.DefineNodeInterface[IConversation]()
var ConversationType = psi.DefineNodeType[*Conversation](psi.WithInterfaceFromNode(ConversationInterface))
var ConversationMessageEdge = psi.DefineEdgeType[*Message]("message")
var _ IConversation = (*Conversation)(nil)

type IConversation interface {
	GetMessages(ctx context.Context) ([]*Message, error)
	SendMessage(ctx context.Context, req *SendMessageRequest) (*Message, error)

	OnMessageReceived(ctx context.Context, req *OnMessageReceivedRequest) error
}

type Conversation struct {
	psi.NodeBase

	Name string `json:"name"`

	Client *openai.Client `json:"-" inject:""`
}

func (c *Conversation) PsiNodeName() string { return c.Name }

func (c *Conversation) GetMessages(ctx context.Context) ([]*Message, error) {
	return psi.GetEdges(c, ConversationMessageEdge), nil
}

func (c *Conversation) SendMessage(ctx context.Context, req *SendMessageRequest) (*Message, error) {
	tx := coreapi.GetTransaction(ctx)

	m, err := c.addMessage(ctx, req)

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

func (c *Conversation) addMessage(ctx context.Context, req *SendMessageRequest) (*Message, error) {
	m := NewMessage()
	m.Timestamp = strconv.FormatInt(time.Now().UnixNano(), 10)
	m.From = req.From
	m.Text = req.Text
	m.Attachments = req.Attachments
	m.Metadata = req.Metadata
	m.SetParent(c)

	if err := m.Update(ctx); err != nil {
		return nil, err
	}

	c.SetEdge(ConversationMessageEdge.Named(m.Timestamp), m)

	return m, nil
}

func (c *Conversation) OnMessageReceived(ctx context.Context, req *OnMessageReceivedRequest) error {
	tx := coreapi.GetTransaction(ctx)
	completionRequest := openai.ChatCompletionRequest{
		TopP:        1.0,
		Temperature: 1.0,
		Model:       "gpt-3.5-turbo",
		MaxTokens:   1000,
		N:           1,
	}

	req.Options.Apply(&completionRequest)

	messages, err := c.GetMessages(ctx)

	if err != nil {
		return err
	}

	lastMessage, err := req.Message.Resolve(ctx)

	if err != nil {
		return err
	}

	appendMessage := func(msg *Message) {
		m := openai.ChatCompletionMessage{}

		m.Name = msg.From.Name
		m.Role = openai.ConvertFromRole(msg.From.Role)
		m.Content = msg.Text

		completionRequest.Messages = append(completionRequest.Messages, m)
	}

	slices.SortFunc(messages, func(i, j *Message) bool {
		return strings.Compare(i.Timestamp, j.Timestamp) < 0
	})

	for _, msg := range messages {
		if msg.Timestamp == lastMessage.Timestamp {
			break
		}
	}

	if len(lastMessage.Attachments) > 0 {
		completionRequest.Messages = append(completionRequest.Messages, openai.ChatCompletionMessage{
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
				def := openai.FunctionDefine{
					Name:        strings.Replace(action.Name, ".", "_", -1),
					Description: action.Description,
					Parameters: &openai.FunctionParams{
						Type:       action.Parameters.Type,
						Required:   action.Parameters.Required,
						Properties: action.Parameters.Properties,
					},
				}

				completionRequest.Functions = append(completionRequest.Functions, &def)
			}

			writer := &bytes.Buffer{}

			if err := rendering.RenderNodeWithTheme(ctx, writer, themes.GlobalTheme, "text/markdown", "", node); err != nil {
				return err
			}

			completionRequest.Messages = append(completionRequest.Messages, openai.ChatCompletionMessage{
				Role: "user",
				Content: fmt.Sprintf(`You have attached a file. Here are the details:
File Path: %s
File Type: %s

%s`, node.CanonicalPath().String(), node.PsiNodeType().Name(), writer.String()),
			})
		}

		completionRequest.FunctionCall = "auto"
	}

	appendMessage(lastMessage)

	result, err := c.Client.CreateChatCompletion(ctx, completionRequest)

	if err != nil {
		return err
	}

	for i, choice := range result.Choices {
		msg := &SendMessageRequest{
			From: UserHandle{
				ID:   completionRequest.Model,
				Name: choice.Message.Name,
				Role: openai.ConvertToRole(choice.Message.Role),
			},

			Text: choice.Message.Content,

			Metadata: map[string]any{
				"oai_choice":   i,
				"oai_request":  completionRequest,
				"oai_response": choice,
			},
		}

		if choice.FinishReason == openai2.FinishReasonFunctionCall {
			msg.Text += "\n\n" + choice.Message.FunctionCall.Name + "\n" + choice.Message.FunctionCall.Arguments
		}

		_, err := c.addMessage(ctx, msg)

		if err != nil {
			return nil
		}
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
	Message *stdlib.Reference[*Message] `json:"message" jsonschema:"title=Message,description=The message received"`
	Options ModelOptions                `json:"options" jsonschema:"title=Model Options,description=The options for the model"`
}
