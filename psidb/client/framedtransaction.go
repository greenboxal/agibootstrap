package client

import (
	"context"
	"time"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"

	rtv1 "github.com/greenboxal/agibootstrap/psidb/apis/rt/v1"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type FramedTransaction struct {
	xid uint64

	driver  *Driver
	builder FrameBuilder

	graph *online.LiveGraph

	open bool
}

func (tx *FramedTransaction) IsOpen() bool             { return tx.open }
func (tx *FramedTransaction) Graph() coreapi.LiveGraph { return tx.graph }

func (tx *FramedTransaction) GetGraphTransaction() coreapi.GraphTransaction {
	panic("implement me")
}

func (tx *FramedTransaction) Add(node psi.Node) {
	tx.graph.Add(node)
}

func (tx *FramedTransaction) Remove(node psi.Node) {
	tx.graph.Remove(node)
}

func (tx *FramedTransaction) Resolve(ctx context.Context, path psi.Path) (psi.Node, error) {
	return tx.graph.ResolveNode(ctx, path)
}

func (tx *FramedTransaction) MakePromise() psi.PromiseHandle {
	return psi.PromiseHandle{
		Xid:   tx.xid,
		Nonce: uint64(time.Now().UnixNano()),
	}
}

func (tx *FramedTransaction) Append(ctx context.Context, entry coreapi.JournalEntry) error {
	tx.builder.Add(entry)

	return nil
}

func (tx *FramedTransaction) Notify(ctx context.Context, not psi.Notification) error {
	if not.Argument != nil {
		if len(not.Params) > 0 {
			return errors.New("notification argument and params are mutually exclusive")
		}

		data, err := ipld.Encode(typesystem.Wrap(not.Argument), dagjson.Encode)

		if err != nil {
			return err
		}

		not.Params = data
		not.Argument = nil
	}

	return tx.Append(ctx, coreapi.JournalEntry{
		Op:           coreapi.JournalOpNotify,
		Notification: &not,
	})
}

func (tx *FramedTransaction) Confirm(ctx context.Context, ack psi.Confirmation) error {
	return tx.Append(ctx, coreapi.JournalEntry{
		Op:           coreapi.JournalOpConfirm,
		Confirmation: &ack,
	})
}

func (tx *FramedTransaction) Wait(ctx context.Context, handles ...psi.Promise) error {
	if len(handles) == 0 {
		return nil
	}

	return tx.Append(ctx, coreapi.JournalEntry{
		Op:       coreapi.JournalOpWait,
		Promises: handles,
	})
}

func (tx *FramedTransaction) Signal(ctx context.Context, handles ...psi.Promise) error {
	if len(handles) == 0 {
		return nil
	}

	return tx.Append(ctx, coreapi.JournalEntry{
		Op:       coreapi.JournalOpSignal,
		Promises: handles,
	})
}

func (tx *FramedTransaction) Commit(ctx context.Context) error {
	if err := tx.ensureOpen(); err != nil {
		return err
	}

	tx.builder.Add(coreapi.JournalEntry{
		Op: coreapi.JournalOpCommit,
	})

	tx.open = false

	frame := tx.builder.Build()

	_, err := tx.driver.PushFrame(ctx, &rtv1.PushFrameRequest{
		Frame: frame,
	})

	if err != nil {
		return err
	}

	return nil
}

func (tx *FramedTransaction) Rollback(ctx context.Context) error {
	if err := tx.ensureOpen(); err != nil {
		return err
	}

	tx.open = false

	return nil
}

func (tx *FramedTransaction) ensureOpen() error {
	if tx.open {
		return nil
	}

	return coreapi.ErrTransactionClosed
}
