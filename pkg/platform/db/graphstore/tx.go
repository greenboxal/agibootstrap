package graphstore

import (
	"context"

	"github.com/greenboxal/agibootstrap/psidb/db/online"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type GraphOperations interface {
	Add(node psi.Node)
	Remove(node psi.Node)

	Resolve(ctx context.Context, path psi.Path) (psi.Node, error)
}

type Transaction interface {
	GraphOperations

	IsOpen() bool
	Graph() *online.LiveGraph
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type transaction struct {
	lg *online.LiveGraph
}

func (t *transaction) IsOpen() bool             { return t.lg.Transaction().IsOpen() }
func (t *transaction) Graph() *online.LiveGraph { return t.lg }

func (t *transaction) Add(node psi.Node) {
	t.lg.Add(node)
}

func (t *transaction) Remove(n psi.Node) {
	t.lg.Remove(n)
}

func (t *transaction) Resolve(ctx context.Context, path psi.Path) (psi.Node, error) {
	return t.lg.ResolveNode(ctx, path)
}

func (t *transaction) Commit(ctx context.Context) error {
	return t.lg.Commit(ctx)
}

func (t *transaction) Rollback(ctx context.Context) error {
	return t.lg.Rollback(ctx)
}

var ctxKeyTransaction = struct{ name string }{name: "PsiDbGsTx"}

func getTransaction(ctx context.Context) *transaction {
	tx, _ := ctx.Value(ctxKeyTransaction).(*transaction)

	return tx
}

func GetTransaction(ctx context.Context) Transaction {
	tx, _ := ctx.Value(ctxKeyTransaction).(Transaction)

	return tx
}

func WithTransaction(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, ctxKeyTransaction, tx)
}
