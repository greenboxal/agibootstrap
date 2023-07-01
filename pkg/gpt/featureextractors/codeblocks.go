package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	mdutils "github.com/greenboxal/agibootstrap/pkg/platform/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type CodeBocks struct {
	psi.NodeBase

	Blocks []mdutils.CodeBlock
}

func ExtractCodeBlocks(ctx context.Context, expectedLanguage string, history ...*thoughtdb.Thought) (*CodeBocks, error) {
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

	cb := &CodeBocks{
		Blocks: blocks,
	}

	cb.Init(cb, "")

	return cb, nil
}
