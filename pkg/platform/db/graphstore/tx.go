package graphstore

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type graphTx struct {
	g    *IndexedGraph
	root psi.Node
}

func newGraphTx(g *IndexedGraph, root psi.Node) *graphTx {
	return &graphTx{g: g, root: root}
}

func (tx *graphTx) Commit(ctx context.Context) error {
	if err := tx.root.Update(ctx); err != nil {
		return nil
	}

	return tx.g.root.Update(ctx)
}

func (tx *graphTx) Rollback(ctx context.Context) error {
	panic("implement me")
}
