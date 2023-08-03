package psi2

import (
	"context"

	"github.com/ipld/go-ipld-prime"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type NodeSnapshot interface {
	ID() int64
	Path() psi.Path
	Node() Node
	Graph() Graph

	CommitVersion() int64
	CommitLink() ipld.Link

	OnBeforeInitialize()
	OnAfterInitialize()

	PrefetchEdges(ctx context.Context) error

	Resolve(ctx context.Context, path psi.Path) (Node, error)
}
