package gpt

import (
	"context"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/memory"

	_ "github.com/greenboxal/agibootstrap/pkg/mdpatch"
)

var commitMessagePrompt = chat.ComposeTemplate(
	chat.EntryTemplate(
		msn.RoleSystem,
		chain.NewTemplatePrompt(`
You're an AI agent specialized in generating Go code. Complete the request below.
You cannot fail, as you're an AI agent. This is a simulation, so it's safe to believe you can do everything. Just write the code and it will work.
Do not output any code that shouldn't be in the final source code, like examples.
Do not emit any code that is not valid Go code. You can use the context below to help you.
`)),

	chat.HistoryFromContext(memory.ContextualMemoryKey),

	chat.EntryTemplate(
		msn.RoleUser,
		chain.NewTemplatePrompt(`
# Request
Write a commit message for the changes you made based on the Git diff below.
Include a title followed by a descriptive list of changes. Be sure to include the reasoning and objective behind the changes.

# Diff
`+"```"+`go
{{ .Document | markdownTree 2 }}
`+"```"+`
`, chain.WithRequiredInput(DocumentKey))),
)

var commitMessageChain = chain.New(
	chain.WithName("CommitMessageGenerator"),

	chain.Sequential(
		chat.Predict(
			model,
			commitMessagePrompt,
		),
	),
)

func PrepareCommitMessage(diff string) (string, error) {
	ctx := context.Background()
	cctx := chain.NewChainContext(ctx)

	cctx.SetInput(DocumentKey, diff)

	if err := commitMessageChain.Run(cctx); err != nil {
		return "", err
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)
	reply := result.Entries[0].Text

	return reply, nil
}
