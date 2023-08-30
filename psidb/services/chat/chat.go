package chat

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
)

type IChat interface {
	GetMessages(ctx context.Context, req *GetMessagesRequest) ([]*Message, error)
	SendMessage(ctx context.Context, req *SendMessageRequest) (*Message, error)
}

var ChatInterface = psi.DefineNodeInterface[IChat]()

type GetMessagesRequest struct {
	From *stdlib.Reference[*Message] `json:"from"`
	To   *stdlib.Reference[*Message] `json:"to"`

	SkipBaseHistory bool `json:"skip_base_history"`
}

type SendMessageRequest struct {
	Timestamp   string         `json:"timestamp" jsonschema:"title=Timestamp,description=The timestamp of the message"`
	From        UserHandle     `json:"from" jsonschema:"title=From,description=The user who sent the message"`
	Text        string         `json:"text" jsonschema:"title=Text,description=The text content of the message"`
	Attachments []psi.Path     `json:"attachments" jsonschema:"title=Attachments,description=The attachments included in the message"`
	Metadata    map[string]any `json:"metadata" jsonschema:"title=Metadata,description=The metadata associated with the message"`

	Function          string `json:"function" jsonschema:"title=Function,description=The function to execute"`
	FunctionArguments string `json:"function_arguments" jsonschema:"title=Function Arguments,description=The arguments for the function"`

	ModelOptions gpt.ModelOptions `json:"model_options" jsonschema:"title=Model Options,description=The options for the model"`
}
