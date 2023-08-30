package agents

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type ThreadContext struct {
	Ctx context.Context

	Client       *openai.Client
	ModelOptions gpt.ModelOptions

	History ChatHistory
	Log     ChatLog

	BaseMessage *chat.Message
}

func (tc *ThreadContext) Fork(ctx context.Context, baseMessage *chat.Message, options gpt.ModelOptions) (*ThreadContext, error) {
	baseHistory := tc.History

	if tc.Log != nil {
		baseHistory = tc.Log
	}

	forked, err := baseHistory.ForkAsChatLog(ctx, baseMessage, options)

	if err != nil {
		return nil, err
	}

	return &ThreadContext{
		Ctx:          tc.Ctx,
		Client:       tc.Client,
		ModelOptions: tc.ModelOptions,
		History:      forked,
		Log:          forked,
	}, nil
}

func (tc *ThreadContext) Merge(ctx context.Context) error {
	if tc.Log == nil {
		return nil
	}

	mergeMsg := chat.NewMessage(chat.MessageKindMerge)

	return tc.Log.AcceptMessage(ctx, mergeMsg)
}

func (tc *ThreadContext) BuildPrompt() *PromptBuilder {
	pb := NewPromptBuilder()
	pb.WithModelOptions(tc.ModelOptions)
	pb.WithClient(tc.Client)
	return pb
}
