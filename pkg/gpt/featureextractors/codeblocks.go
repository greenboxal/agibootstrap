package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt"
	mdutils2 "github.com/greenboxal/agibootstrap/pkg/platform/mdutils"
)

type CodeBocks struct {
	Blocks []mdutils2.CodeBlock
}

func ExtractCodeBlocks(ctx context.Context, expectedLanguage string, history ...agents.Message) (CodeBocks, error) {
	var blocks []mdutils2.CodeBlock

	for _, msg := range history {
		raw := msg.Text

		if expectedLanguage != "" {
			raw = gpt.SanitizeCodeBlockReply(raw, expectedLanguage)
		}

		node := mdutils2.ParseMarkdown([]byte(raw))
		b := mdutils2.ExtractCodeBlocks(node)

		blocks = append(blocks, b...)
	}

	return CodeBocks{
		Blocks: blocks,
	}, nil
}
