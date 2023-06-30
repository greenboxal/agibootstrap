package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
)

type Text struct {
	LeafBase

	Text string
}

func (l *Text) Render(ctx context.Context, tb *rendering.TokenBuffer) error {
	_, err := tb.Write([]byte(l.Text))

	return err
}
