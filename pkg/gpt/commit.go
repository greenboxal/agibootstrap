package gpt

import (
	"context"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/memory"

	"github.com/greenboxal/agibootstrap/pkg/mdutils"
)

var CommitMessagePrompt = chat.ComposeTemplate(
	chat.EntryTemplate(
		msn.RoleSystem,
		chain.NewTemplatePrompt(`
You're an AI agent specialized in generation good summaries for git diffs and commits, and synthesizing commit messages. Complete the request below.
`)),

	chat.HistoryFromContext(memory.ContextualMemoryKey),

	chat.EntryTemplate(
		msn.RoleUser,
		chain.NewTemplatePrompt(`
Write a really good commit message for the changes you made based on the Git diff below.

## Examples
`+"```markdown"+`
Implement feature X, Y, Z

- Add feature X
- Add feature Y
- Add feature Z
`+"```"+`

`+"```markdown"+`
Refactor feature 

`+"```"+`
`)),

	FunctionCallRequest("Human", "getCommitMessage", RequestKey),

	chat.EntryTemplate(
		msn.RoleAI,
		chain.NewTemplatePrompt("\t```markdown")),
)

var CommitMessageChain = chain.New(
	chain.WithName("CommitMessageGenerator"),

	chain.Sequential(
		chat.Predict(
			GlobalModel,
			CommitMessagePrompt,
			chat.WithMaxTokens(1024),
		),
	),
)

func PrepareCommitMessage(diff string) (string, error) {
	ctx := context.Background()
	cctx := chain.NewChainContext(ctx)

	req := ContextBag{}
	req["git diff --cached"] = mdutils.CodeBlock{
		Filename: "",
		Language: "diff",
		Code:     diff,
	}

	cctx.SetInput(RequestKey, req)

	if err := CommitMessageChain.Run(cctx); err != nil {
		return "", err
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)
	reply := result.Entries[0].Text

	return reply, nil
}
