package graphfs

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Transaction struct {
	mu      sync.Mutex
	journal *Journal
	txm     *TransactionManager
	xid     uint64
	log     []*JournalEntry
	done    bool

	isReadOnly bool

	dirtyNodes map[int64]*txNode
}

func (tx *Transaction) GetXid() uint64          { return tx.xid }
func (tx *Transaction) GetLog() []*JournalEntry { return tx.log }
func (tx *Transaction) IsOpen() bool            { return !tx.done }

func (tx *Transaction) Notify(ctx context.Context, not psi.Notification) error {
	return tx.Append(ctx, JournalEntry{
		Op:           JournalOpNotify,
		Notification: &not,
	})
}

func (tx *Transaction) Append(ctx context.Context, entry JournalEntry) error {
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

	if err := tx.append(JournalEntry{
		Op: JournalOpRollback,
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

func (tx *Transaction) append(entry JournalEntry) error {
	if len(tx.log) == 0 && entry.Op != JournalOpBegin {
		if err := tx.append(JournalEntry{
			Op: JournalOpBegin,
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
	case JournalOpBegin:
		tx.xid = entry.Xid

	case JournalOpWrite:
		n := tx.getStagedNode(entry.Inode)
		n.Frozen = *entry.Node

	case JournalOpSetEdge:
		n := tx.getStagedNode(entry.Inode)
		n.Edges[entry.Edge.Key.String()] = *entry.Edge

	case JournalOpRemoveEdge:
		k := entry.Edge.Key.String()
		n := tx.getStagedNode(entry.Inode)
		e, ok := n.Edges[k]

		if !ok {
			e = *entry.Edge
		}

		e.Flags |= EdgeFlagRemoved

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
		Edges: map[string]SerializedEdge{},
	}

	tx.dirtyNodes[ino] = n

	return n
}

var ctxKeyTransaction = &struct{}{}

func WithTransaction(ctx context.Context, tx *Transaction) context.Context {
	return context.WithValue(ctx, ctxKeyTransaction, tx)
}

func GetTransaction(ctx context.Context) *Transaction {
	tx, _ := ctx.Value(ctxKeyTransaction).(*Transaction)

	return tx
}
