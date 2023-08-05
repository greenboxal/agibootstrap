package client

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
)

type transaction struct {
	client *Client
	lg     *online.LiveGraph
	opts   coreapi.TransactionOptions
}

func (t *transaction) IsOpen() bool                          { return t.lg.Transaction().IsOpen() }
func (t *transaction) Graph() *online.LiveGraph              { return t.lg }
func (t *transaction) ServiceLocator() inject.ServiceLocator { return t }

func (t *transaction) GetService(key inject.ServiceKey) (any, error) {
	if sl := t.opts.ServiceLocator; sl != nil {
		r, err := sl.GetService(key)

		if err == nil {
			return r, nil
		} else if err != inject.ServiceNotFound {
			return nil, err
		}
	}

	return t.client.sp.GetService(key)
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
	if err := t.lg.Commit(ctx); err != nil {
		return err
	}

	return t.Close()
}

func (t *transaction) Rollback(ctx context.Context) error {
	if err := t.lg.Rollback(ctx); err != nil {
		return err
	}

	return t.Close()
}

func (t *transaction) Close() error {
	if t.IsOpen() {
		if err := t.Rollback(context.Background()); err != nil {
			return err
		}
	}

	return t.lg.Close()
}
