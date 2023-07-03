package promptml

import (
	"bytes"
	"context"
	"io"

	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
)

type Stage struct {
	MaxTokens int
	Tokenizer tokenizers.BasicTokenizer

	root Parent
}

func NewStage(root Parent, tokenizer tokenizers.BasicTokenizer) *Stage {
	s := &Stage{
		Tokenizer: tokenizer,
	}

	s.setRoot(root)

	return s
}

func (s *Stage) Render(ctx context.Context, writer io.Writer) error {
	return s.RenderNode(ctx, s.root, writer)
}

func (s *Stage) RenderNode(ctx context.Context, node Parent, writer io.Writer) error {
	for !s.root.IsValid() || s.root.NeedsLayout() || !node.IsValid() || node.NeedsLayout() {
		if err := node.Update(ctx); err != nil {
			return err
		}

		if err := s.root.Layout(ctx); err != nil {
			return err
		}
	}

	_, err := node.PmlNodeBase().GetTokenBuffer().WriteTo(writer)

	if err != nil {
		return err
	}

	return nil
}

func (s *Stage) RenderNodeToString(ctx context.Context, node Parent) (string, error) {
	buf := bytes.NewBuffer(nil)

	if err := s.RenderNode(ctx, node, buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *Stage) RenderToString(ctx context.Context) (string, error) {
	return s.RenderNodeToString(ctx, s.root)
}

func (s *Stage) setRoot(root Parent) {
	if s.root == root {
		return
	}

	s.root = root

	s.root.PmlNodeBase().setStage(s)
}
