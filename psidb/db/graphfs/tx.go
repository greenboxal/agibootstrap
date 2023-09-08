package graphfs

import (
	"context"
	"sync"
	"time"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type Transaction struct {
	mu      sync.Mutex
	journal coreapi.Journal
	txm     *TransactionManager
	xid     uint64
	log     []*coreapi.JournalEntry
	done    bool

	isReadOnly bool

	dirtyNodes map[int64]*txNode
}

func (tx *Transaction) GetXid() uint64                  { return tx.xid }
func (tx *Transaction) GetLog() []*coreapi.JournalEntry { return tx.log }
func (tx *Transaction) IsOpen() bool                    { return !tx.done }

func (tx *Transaction) Notify(ctx context.Context, not psi.Notification) error {
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

func (tx *Transaction) Confirm(ctx context.Context, ack psi.Confirmation) error {
	return tx.Append(ctx, coreapi.JournalEntry{
		Op:           coreapi.JournalOpConfirm,
		Confirmation: &ack,
	})
}

func (tx *Transaction) Wait(ctx context.Context, handles ...psi.Promise) error {
	if len(handles) == 0 {
		return nil
	}

	return tx.Append(ctx, coreapi.JournalEntry{
		Op:       coreapi.JournalOpWait,
		Promises: handles,
	})
}

func (tx *Transaction) Signal(ctx context.Context, handles ...psi.Promise) error {
	if len(handles) == 0 {
		return nil
	}

	return tx.Append(ctx, coreapi.JournalEntry{
		Op:       coreapi.JournalOpSignal,
		Promises: handles,
	})
}

func (tx *Transaction) Append(ctx context.Context, entry coreapi.JournalEntry) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.done {
		panic(errors.New("transaction already finished"))
	}

	return tx.append(entry)
}

func (tx *Transaction) Commit(ctx context.Context) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.done {
		return errors.New("transaction already finished")
	}

	if len(tx.log) == 0 {
		return tx.close()
	}

	if err := tx.txm.commitTransaction(ctx, tx); err != nil {
		return err
	}

	return tx.close()
}

func (tx *Transaction) Rollback(ctx context.Context) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.done {
		return errors.New("transaction already finished")
	}

	if err := tx.append(coreapi.JournalEntry{
		Op: coreapi.JournalOpRollback,
	}); err != nil {
		return err
	}

	return tx.close()
}

func (tx *Transaction) close() error {
	if tx.done {
		return nil
	}

	tx.done = true

	tx.txm.updateTransaction(tx)

	return nil
}

func (tx *Transaction) append(entry coreapi.JournalEntry) error {
	if len(tx.log) == 0 && entry.Op != coreapi.JournalOpBegin {
		if err := tx.append(coreapi.JournalEntry{
			Op: coreapi.JournalOpBegin,
		}); err != nil {
			return err
		}
	}

	entry.Xid = tx.xid

	if entry.Ts == 0 {
		entry.Ts = time.Now().UnixNano()
	}

	if tx.journal != nil && !tx.isReadOnly {
		if err := tx.journal.Write(&entry); err != nil {
			return err
		}
	}

	switch entry.Op {
	case coreapi.JournalOpBegin:
		tx.xid = entry.Xid

	case coreapi.JournalOpWrite:
		n := tx.getStagedNode(entry.Inode)
		n.Frozen = *entry.Node

	case coreapi.JournalOpSetEdge:
		n := tx.getStagedNode(entry.Inode)
		n.Edges[entry.Edge.Key.String()] = *entry.Edge

	case coreapi.JournalOpRemoveEdge:
		k := entry.Edge.Key.String()
		n := tx.getStagedNode(entry.Inode)
		e, ok := n.Edges[k]

		if !ok {
			e = *entry.Edge
		}

		e.Flags |= coreapi.EdgeFlagRemoved

		n.Edges[k] = e
	}

	tx.log = append(tx.log, &entry)

	return nil
}

func (tx *Transaction) getStagedNode(ino int64) *txNode {
	if n := tx.dirtyNodes[ino]; n != nil {
		return n
	}

	n := &txNode{
		Inode: ino,
		Edges: map[string]coreapi.SerializedEdge{},
	}

	tx.dirtyNodes[ino] = n

	return n
}
