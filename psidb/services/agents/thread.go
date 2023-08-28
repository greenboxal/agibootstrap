package agents

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
)

type ThreadContext struct {
	Ctx context.Context

	Client       *openai.Client
	ModelOptions ModelOptions

	History ChatHistory
	Log     ChatLog

	BaseMessage *Message
}

func (tc *ThreadContext) Fork(ctx context.Context, baseMessage *Message, options ModelOptions) (*ThreadContext, error) {
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

	mergeMsg := NewMessage(MessageKindMerge)

	return tc.Log.AcceptMessage(ctx, mergeMsg)
}

func (tc *ThreadContext) BuildPrompt() *PromptBuilder {
	pb := NewPromptBuilder()
	pb.WithModelOptions(tc.ModelOptions)
	pb.WithClient(tc.Client)
	return pb
}
