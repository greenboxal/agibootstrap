package core

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
)

type transaction struct {
	core *Core
	lg   *online.LiveGraph
}

func (t *transaction) IsOpen() bool             { return t.lg.Transaction().IsOpen() }
func (t *transaction) Graph() *online.LiveGraph { return t.lg }

func (t *transaction) GetService(key inject.ServiceKey) (any, error) {
	return t.core.sp.GetService(key)
}

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

var ctxKeyTransaction = &struct{}{}

func getTransaction(ctx context.Context) *transaction {
	tx, _ := ctx.Value(ctxKeyTransaction).(*transaction)

	return tx
}

func GetTransaction(ctx context.Context) coreapi.Transaction {
	tx, _ := ctx.Value(ctxKeyTransaction).(coreapi.Transaction)

	return tx
}

func WithTransaction(ctx context.Context, tx coreapi.Transaction) context.Context {
	return context.WithValue(ctx, ctxKeyTransaction, tx)
}
