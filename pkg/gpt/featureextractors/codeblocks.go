package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/agents"
	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/mdutils"
)

type CodeBocks struct {
	Blocks []mdutils.CodeBlock
}

func ExtractCodeBlocks(ctx context.Context, expectedLanguage string, history... agents.Message) (CodeBocks, error) {
	var blocks []mdutils.CodeBlock

	for _, msg := range history {
		raw := msg.Text

		if expectedLanguage != "" {
			raw = gpt.SanitizeCodeBlockReply(raw, expectedLanguage)
		}

		node := mdutils.ParseMarkdown([]byte(raw))
		b := mdutils.ExtractCodeBlocks(node)

		blocks = append(blocks, b...)
	}

	return CodeBocks{
		Blocks: blocks,
	}, nil
}
