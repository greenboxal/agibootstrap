package core

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.opentelemetry.io/otel"
	propagation2 "go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
)

type transaction struct {
	core    *Core
	session coreapi.Session
	lg      *online.LiveGraph

	opts coreapi.TransactionOptions
}

func (t *transaction) IsOpen() bool                                  { return t.lg.Transaction().IsOpen() }
func (t *transaction) Graph() coreapi.LiveGraph                      { return t.lg }
func (t *transaction) GetGraphTransaction() coreapi.GraphTransaction { return t.lg.Transaction() }
func (t *transaction) ServiceLocator() inject.ServiceLocator         { return t }

func (t *transaction) GetService(key inject.ServiceKey) (any, error) {
	if sl := t.opts.ServiceLocator; sl != nil {
		r, err := sl.GetService(key)

		if err == nil {
			return r, nil
		} else if !errors.Is(err, inject.ServiceNotFound) {
			return nil, err
		}
	}

	return t.core.serviceProvider.GetService(key)
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

func (t *transaction) MakePromise() psi.PromiseHandle {
	return psi.PromiseHandle{
		Xid:   t.lg.Transaction().GetXid(),
		Nonce: uint64(time.Now().UnixNano()),
	}
}

func (t *transaction) Notify(ctx context.Context, not psi.Notification) error {
	if not.Nonce == 0 {
		not.Nonce = uint64(time.Now().UnixNano())
	}

	if not.SessionID == "" && t.session != nil {
		not.SessionID = t.session.UUID()
	}

	if span := trace.SpanContextFromContext(ctx); span.IsValid() {
		carrier := propagation2.MapCarrier{}

		otel.GetTextMapPropagator().Inject(ctx, carrier)

		data, err := json.Marshal(carrier)

		if err != nil {
			return err
		}

		not.TraceID = string(data)
		/*&psi.NotificationTrace{
			TraceID:    span.TraceID().String(),
			SpanID:     span.SpanID().String(),
			TraceFlags: int(span.TraceFlags()),
			TraceState: span.TraceState().String(),
			Remote:     span.IsRemote(),
		}*/
	}

	return t.lg.Transaction().Notify(ctx, not)
}

func (t *transaction) Confirm(ctx context.Context, ack psi.Confirmation) error {
	return t.lg.Transaction().Confirm(ctx, ack)
}

func (t *transaction) Wait(ctx context.Context, handles ...psi.Promise) error {
	return t.lg.Transaction().Wait(ctx, handles...)
}

func (t *transaction) Signal(ctx context.Context, handles ...psi.Promise) error {
	return t.lg.Transaction().Signal(ctx, handles...)
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
