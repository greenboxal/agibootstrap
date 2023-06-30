package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
)

type Leaf interface {
	Node

	Render(ctx context.Context, tb *rendering.TokenBuffer) error
}

type LeafBase struct {
	NodeBase

	tb            *rendering.TokenBuffer
	isRenderValid bool
}

func (l *LeafBase) Update(ctx context.Context) error {
	if !l.isRenderValid {
		l.tb.Reset()

		if err := l.NodeBase.Update(ctx); err != nil {
			return err
		}

		if err := l.Render(ctx, l.tb); err != nil {
			return err
		}

		l.currentLength = l.tb.TokenCount()
		l.isRenderValid = true

		l.RequestParentLayout()
	}

	return nil
}

func (l *LeafBase) InvalidateRender() {
	l.isRenderValid = false

	l.Invalidate()
}

func (l *LeafBase) Render(ctx context.Context, tb *rendering.TokenBuffer) error {
	return nil
}
