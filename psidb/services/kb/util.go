package kb

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/gpt/featureextractors"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
)

func runChainWithMessages(ctx context.Context, msgs []*thoughtdb.Thought) (string, error) {
	prompt := &featureextractors.SimplePromptTemplate{Messages: msgs}
	cctx := chain.NewChainContext(ctx)

	stepChain := chain.New(
		chain.WithName("DocumentSummarizer"),

		chain.Sequential(
			chat.Predict(
				gpt.GlobalModel,
				prompt,
				chat.WithMaxTokens(1024),
			),
		),
	)

	if err := stepChain.Run(cctx); err != nil {
		return "", err
	}

	reply := chain.Output(cctx, chat.ChatReplyContextKey)

	if len(reply.Entries) == 0 {
		return "", nil
	}

	return reply.Entries[0].Text, nil
}
