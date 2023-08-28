package stdlib

import (
	"context"
	"reflect"
	"sync/atomic"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Reference[T psi.Node] struct {
	psi.Path

	cached         T
	hasCachedValue atomic.Bool
}

func RefFromPath[T psi.Node](p psi.Path) *Reference[T] {
	return &Reference[T]{Path: p}
}

func Ref[T psi.Node](node T) *Reference[T] {
	v := reflect.ValueOf(node)

	if !v.IsValid() || v.IsNil() {
		return nil
	}

	p := node.CanonicalPath()
	return &Reference[T]{Path: p}
}

func (nr *Reference[T]) IsEmpty() bool {
	if nr == nil {
		return true
	}

	return nr.Path.IsEmpty()
}

func (nr *Reference[T]) Get(ctx context.Context) (_ T) {
	if nr == nil {
		return
	}

	v, err := nr.Resolve(ctx)

	if err != nil {
		panic(err)
	}

	return v
}

func (nr *Reference[T]) Resolve(ctx context.Context) (empty T, _ error) {
	if nr == nil || nr.IsEmpty() {
		return empty, psi.ErrNodeNotFound
	}

	if !nr.hasCachedValue.Load() {
		tx := coreapi.GetTransaction(ctx)
		v, err := psi.Resolve[T](ctx, tx.Graph(), nr.Path)

		if err != nil {
			return nr.cached, err
		}

		nr.cached = v
		nr.hasCachedValue.Store(true)
	}

	return nr.cached, nil
}
