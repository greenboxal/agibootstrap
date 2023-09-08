package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering"
)

type Leaf interface {
	Node

	Render(ctx context.Context, tb *rendering.TokenBuffer) error

	GetTokenBuffer() *rendering.TokenBuffer
}

type LeafBase struct {
	NodeBase
}

func (l *LeafBase) Init(self psi.Node) {
	l.NodeBase.Init(self)
}

func (l *LeafBase) PmlLeaf() Leaf { return l.PsiNode().(Leaf) }

func (l *LeafBase) OnUpdate(ctx context.Context) error {
	if err := l.NodeBase.OnUpdate(ctx); err != nil {
		return err
	}

	if l.tb != nil {
		l.tb.Reset()

		if err := l.render(ctx, l.tb); err != nil {
			return err
		}

		l.tokenLength.SetValue(l.tb.TokenCount())
	}

	/*if l.GetTokenLength() > l.GetEffectiveMaxLength() && l.PmlNode().IsResizable() {
		l.SetVisible(false)
	} else {
		l.SetVisible(true)
	}*/

	return nil
}

func (l *LeafBase) Render(ctx context.Context, tb *rendering.TokenBuffer) error {
	return nil
}

func (l *LeafBase) render(ctx context.Context, tb *rendering.TokenBuffer) error {
	if err := l.PmlLeaf().Render(ctx, tb); err != nil {
		return err
	}

	return nil
}
