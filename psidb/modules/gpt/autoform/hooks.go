package autoform

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
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
