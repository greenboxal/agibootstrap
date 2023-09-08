package session

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel"
	propagation2 "go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type transaction struct {
	core    coreapi.Core
	session *Session
	lg      *online.LiveGraph
	sp      inject.ServiceProvider

	opts coreapi.TransactionOptions
}

func (t *transaction) IsOpen() bool                                  { return t.lg.Transaction().IsOpen() }
func (t *transaction) Graph() coreapi.LiveGraph                      { return t.lg }
func (t *transaction) GetGraphTransaction() coreapi.GraphTransaction { return t.lg.Transaction() }
func (t *transaction) ServiceLocator() inject.ServiceLocator         { return t.sp }

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
