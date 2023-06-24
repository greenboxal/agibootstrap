package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	mdutils2 "github.com/greenboxal/agibootstrap/pkg/platform/mdutils"
)

type CodeBocks struct {
	Blocks []mdutils2.CodeBlock
}

func ExtractCodeBlocks(ctx context.Context, expectedLanguage string, history ...thoughtstream.Thought) (CodeBocks, error) {
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
