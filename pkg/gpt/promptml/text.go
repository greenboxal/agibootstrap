package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
)

type TextNode struct {
	LeafBase

	Text obsfx.StringProperty `json:"text"`
}

func (l *TextNode) Init(self psi.Node) {
	l.LeafBase.Init(self)

	obsfx.ObserveInvalidation(&l.Text, l.Invalidate)
}

func (l *TextNode) Render(ctx context.Context, tb *rendering.TokenBuffer) error {
	_, err := tb.Write([]byte(l.Text.Value()))

	return err
}

func Text(content string) *TextNode {
	t := &TextNode{}

	t.Init(t)

	t.Text.SetValue(content)

	return t
}

func TextWithData(binding obsfx.ObservableValue[string]) *TextNode {
	t := &TextNode{}

	t.Init(t)

	t.Text.Bind(binding)

	return t
}
