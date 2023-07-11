package featureextractors

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	mdutils2 "github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

type CodeBocks struct {
	psi.NodeBase

	Blocks []mdutils2.CodeBlock
}

func ExtractCodeBlocks(ctx context.Context, expectedLanguage string, history ...*thoughtdb.Thought) (*CodeBocks, error) {
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

	cb := &CodeBocks{
		Blocks: blocks,
	}

	cb.Init(cb)

	return cb, nil
}
