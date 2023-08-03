package psi2

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type NodeReference[T Node] interface {
	IsLoaded() bool

	Resolve(ctx context.Context) (T, error)
	ResolveSync() T
}

type LazyNodeLoadFunc[T Node] func(ctx context.Context) (T, error)

type LazyNodeReference[T Node] struct {
	loader LazyNodeLoadFunc[T]
	loaded bool
	node   T
	err    error
}

func NewLazyNodeReference[T Node](loader LazyNodeLoadFunc[T]) *LazyNodeReference[T] {
	return &LazyNodeReference[T]{
		loader: loader,
	}
}

func (l *LazyNodeReference[T]) IsLoaded() bool { return l.loaded }

func (l *LazyNodeReference[T]) ResolveSync() T {
	n, err := l.Resolve(nil)

	if err != nil {
		panic(err)
	}

	return n
}

func (l *LazyNodeReference[T]) Resolve(ctx context.Context) (empty T, _ error) {
	if !l.loaded {
		if err := l.Load(ctx); err != nil {
			return empty, err
		}
	}

	if l.err != nil {
		return empty, l.err
	}

	return l.node, l.err
}

func (l *LazyNodeReference[T]) Invalidate() {
	l.loaded = false
}

func (l *LazyNodeReference[T]) Load(ctx context.Context) error {
	node, err := l.loader(ctx)

	if err != nil {
		l.err = err
	} else {
		l.node = node
	}

	l.loaded = true

	return err
}

type StaticNodeReference[T Node] struct {
	node T
}

func NewStaticNodeReference[T Node](node T) *StaticNodeReference[T] {
	return &StaticNodeReference[T]{
		node: node,
	}
}

func (s *StaticNodeReference[T]) IsLoaded() bool { return true }

func (s *StaticNodeReference[T]) Resolve(ctx context.Context) (T, error) { return s.node, nil }
func (s *StaticNodeReference[T]) ResolveSync() T                         { return s.node }

type MutableNodeReference[T Node] struct {
	obsfx.SimpleProperty[NodeReference[T]]
}

func (m *MutableNodeReference[T]) IsLoaded() bool {
	r := m.Value()

	return r != nil && r.IsLoaded()
}

func (m *MutableNodeReference[T]) Resolve(ctx context.Context) (empty T, _ error) {
	r := m.Value()

	if r == nil {
		return empty, psi.ErrNodeNotFound
	}

	return r.Resolve(ctx)
}

func (m *MutableNodeReference[T]) ResolveSync() (empty T) {
	r := m.Value()

	if r == nil {
		return empty
	}

	return r.ResolveSync()
}

type EmptyNodeReference[T Node] struct{}

func (e *EmptyNodeReference[T]) IsLoaded() bool                               { return true }
func (e *EmptyNodeReference[T]) Resolve(ctx context.Context) (def T, _ error) { return def, nil }
func (e *EmptyNodeReference[T]) ResolveSync() (def T)                         { return def }

func NewRelativeNodeReference[T Node](relativeTo Node, path psi.Path) NodeReference[T] {
	return NewLazyNodeReference(func(ctx context.Context) (empty T, _ error) {
		resolvedAny, err := relativeTo.PsiNodeSnapshot().Resolve(ctx, path)

		if err != nil {
			return empty, err
		}

		resolved, ok := resolvedAny.(T)

		if !ok {
			return empty, fmt.Errorf("resolved node is not of type %T", empty)
		}

		return resolved, nil
	})
}

func ResolveLoadedOrNil[T Node](ref NodeReference[T]) (empty T) {
	if ref == nil {
		return empty
	}

	if !ref.IsLoaded() {
		return empty
	}

	return ref.ResolveSync()
}
