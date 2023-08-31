package coreapi

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/otel"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type GraphOperations interface {
	Add(node psi.Node)
	Remove(node psi.Node)

	Resolve(ctx context.Context, path psi.Path) (psi.Node, error)
}

type TransactionOption func(*TransactionOptions)

type TransactionOptions struct {
	ReadOnly       bool
	ServiceLocator inject.ServiceLocator
}

func (o *TransactionOptions) Apply(options ...TransactionOption) {
	for _, option := range options {
		option(o)
	}
}

func WithReadOnly() TransactionOption {
	return func(o *TransactionOptions) {
		o.ReadOnly = true
	}
}

func WithServiceLocator(sl inject.ServiceLocator) TransactionOption {
	return func(o *TransactionOptions) {
		o.ServiceLocator = sl
	}
}

type LiveGraph interface {
	psi.Graph

	GraphOperations

	Delete(ctx context.Context, node psi.Node) error

	Root() psi.Path
	Services() inject.ServiceLocator
}

type Transaction interface {
	GraphOperations

	IsOpen() bool
	Graph() LiveGraph

	MakePromise() psi.PromiseHandle

	Notify(ctx context.Context, not psi.Notification) error
	Confirm(ctx context.Context, ack psi.Confirmation) error
	Wait(ctx context.Context, handles ...psi.Promise) error
	Signal(ctx context.Context, handles ...psi.Promise) error

	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TransactionFunc func(ctx context.Context, tx Transaction) error

type TransactionOperations interface {
	BeginTransaction(ctx context.Context, options ...TransactionOption) (Transaction, error)
	RunTransaction(ctx context.Context, fn TransactionFunc, options ...TransactionOption) (err error)
}

var tracer = otel.Tracer("coreapi")

func RunTransaction(
	ctx context.Context,
	ops TransactionOperations,
	fn TransactionFunc,
	options ...TransactionOption,
) (err error) {
	tx := GetTransaction(ctx)

	ctx, span := tracer.Start(ctx, "coreapi.RunTransaction")
	defer span.End()

	if tx == nil {
		tx, err = ops.BeginTransaction(ctx, options...)

		if err != nil {
			return err
		}

		defer func() {
			if e := recover(); e != nil {
				er := fmt.Errorf("%v", e)

				if tx.IsOpen() {
					if err := tx.Rollback(ctx); err != nil {
						panic(multierror.Append(err, er))
					}
				}

				panic(e)
			}
		}()
	}

	ctx = WithTransaction(ctx, tx)
	err = fn(ctx, tx)

	if tx.IsOpen() {
		if err != nil {
			if e := tx.Rollback(ctx); e != nil {
				err = multierror.Append(err, e)
			}
		} else {
			if e := tx.Commit(ctx); e != nil {
				err = multierror.Append(err, e)
			}
		}
	}

	return
}

var ctxKeyTransaction = struct{ name string }{name: "PsiDbTx"}

func GetTransaction(ctx context.Context) Transaction {
	tx := ctx.Value(ctxKeyTransaction)

	if tx == nil {
		return nil
	}

	return tx.(Transaction)
}

func WithTransaction(ctx context.Context, tx Transaction) context.Context {
	return context.WithValue(ctx, ctxKeyTransaction, tx)
}

func Dispatch(ctx context.Context, not psi.Notification, options ...psi.NotificationOption) error {
	tx := GetTransaction(ctx)

	if tx == nil {
		return fmt.Errorf("no transaction")
	}

	if len(options) > 0 {
		not = not.WithOptions(options...)
	}

	return tx.Notify(ctx, not)
}
