package stdlib

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Reference[T psi.Node] struct {
	Path *psi.Path `json:"path"`
}

func Ref[T psi.Node](node T) *Reference[T] {
	p := node.CanonicalPath()
	return &Reference[T]{Path: &p}
}

func (nr Reference[T]) Resolve(ctx context.Context) (T, error) {
	tx := coreapi.GetTransaction(ctx)

	return psi.Resolve[T](ctx, tx.Graph(), *nr.Path)
}
