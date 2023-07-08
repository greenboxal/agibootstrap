package psi

import (
	"context"

	"github.com/pkg/errors"
)

type EdgeTypeOption func(*edgeType)

func WithEdgeTypeName(name string) EdgeTypeOption {
	return func(et *edgeType) {
		et.name = name
	}
}

func WithEdgeTypeVirtual() EdgeTypeOption {
	return func(et *edgeType) {
		et.virtual = true
	}
}

func WithEdgeTypeIndexed() EdgeTypeOption {
	return func(et *edgeType) {
		et.indexed = true
	}
}

func WithEdgeTypeNamed() EdgeTypeOption {
	return func(et *edgeType) {
		et.named = true
	}
}

func WithEdgeTypeResolveFunc(fn ResolveEdgeFunc) EdgeTypeOption {
	return func(et *edgeType) {
		et.resolve = fn
	}
}

type EdgeType interface {
	Kind() EdgeKind

	IsVirtual() bool
	IsIndexed() bool
	IsNamed() bool

	Resolve(ctx context.Context, g Graph, from Node, key EdgeKey) (Node, error)
}

type ResolveEdgeFunc func(ctx context.Context, g Graph, from Node, key EdgeKey) (Node, error)

type edgeType struct {
	kind EdgeKind

	name string

	virtual bool
	indexed bool
	named   bool

	resolve ResolveEdgeFunc
}

func (t *edgeType) Kind() EdgeKind  { return t.kind }
func (t *edgeType) IsVirtual() bool { return t.virtual }
func (t *edgeType) IsIndexed() bool { return t.indexed }
func (t *edgeType) IsNamed() bool   { return t.named }

func (t *edgeType) Resolve(ctx context.Context, g Graph, from Node, key EdgeKey) (Node, error) {
	if t.resolve != nil {
		return t.resolve(ctx, g, from, key)
	}

	return nil, errors.New("edge type does not support resolve")
}
