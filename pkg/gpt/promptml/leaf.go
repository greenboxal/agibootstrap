package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
)

type Leaf interface {
	Node

	Render(ctx context.Context) error

	GetTokenBuffer() *rendering.TokenBuffer
}

type LeafBase struct {
	NodeBase
}

func (l *LeafBase) PmlLeaf() Leaf { return l.PsiNode().(Leaf) }

func (l *LeafBase) Update(ctx context.Context) error {
	if err := l.NodeBase.Update(ctx); err != nil {
		return err
	}

	l.tb.Reset()

	if err := l.PmlLeaf().Render(ctx); err != nil {
		return err
	}

	l.RequestParentLayout()

	return nil
}

func (l *LeafBase) Render(ctx context.Context) error {
	return nil
}
