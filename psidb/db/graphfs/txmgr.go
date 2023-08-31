package graphfs

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
)

type TransactionManager struct {
	logger     *otelzap.SugaredLogger
	journal    *Journal
	checkpoint coreapi.Checkpoint
	graph      *VirtualGraph

	mu                 sync.RWMutex
	activeTransactions map[uint64]*Transaction

	closed bool
}

func NewTransactionManager(graph *VirtualGraph, journal *Journal, checkpoint coreapi.Checkpoint) *TransactionManager {
	return &TransactionManager{
		logger:             logging.GetLogger("graphfs/transactionManager"),
		graph:              graph,
		journal:            journal,
		checkpoint:         checkpoint,
		activeTransactions: make(map[uint64]*Transaction),
	}
}

type notificationKey struct {
	xid   uint64
	rid   uint64
	nonce uint64
}

func (txm *TransactionManager) Recover(ctx context.Context) error {
	xid, ok, err := txm.checkpoint.Get()

	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	if xid+1 < txm.journal.nextIndex {
		txm.logger.Infow("Recovering recent transactions", "from", xid, "to", txm.journal.nextIndex)
	}

	if err != nil {
		return err
	}

	it := txm.journal.Iterate(xid+1, -1)

	notifications := map[notificationKey]*coreapi.JournalEntry{}
	recoveredTransactions := map[uint64]*Transaction{}

	for it.Next() {
		var tx *Transaction

		entry := it.Value()

		if entry.Op == coreapi.JournalOpBegin {
			tx = txm.newTransaction(true)

			recoveredTransactions[entry.Xid] = tx
		} else {
			tx = recoveredTransactions[entry.Xid]
		}

		if tx == nil {
			return errors.New("invalid transaction id")
		}

		if tx.done {
			continue
		}

		if err := tx.Append(ctx, entry); err != nil {
			return err
		}

		if entry.Op == coreapi.JournalOpCommit {
			if err := txm.commitTransaction(ctx, tx); err != nil {
				return err
			}
		} else if entry.Op == coreapi.JournalOpRollback {
			if err := tx.Rollback(ctx); err != nil {
				return err
			}
		} else if entry.Op == coreapi.JournalOpNotify {
			notifications[notificationKey{
				xid:   entry.Xid,
				rid:   entry.Rid,
				nonce: entry.Notification.Nonce,
			}] = &entry
		} else if entry.Op == coreapi.JournalOpConfirm {
			not := notifications[notificationKey{
				xid:   entry.Confirmation.Xid,
				rid:   entry.Confirmation.Rid,
				nonce: entry.Confirmation.Nonce,
			}]

			if not != nil {
				not.Confirmation = entry.Confirmation
			}
		}
	}

	for _, tx := range recoveredTransactions {
		if !tx.done {
			if err := tx.Rollback(ctx); err != nil {
				txm.logger.Error(err)
			}
		}
	}

	return nil
}

func (txm *TransactionManager) newTransaction(isReadOnly bool) *Transaction {
	tx := &Transaction{
		txm:        txm,
		dirtyNodes: make(map[int64]*txNode),
	}

	if !isReadOnly {
		tx.journal = txm.journal
	}

	return tx
}

func (txm *TransactionManager) BeginTransaction(ctx context.Context) (*Transaction, error) {
	if txm.closed {
		return nil, errors.New("transaction manager already closed")
	}

	tx := txm.newTransaction(false)

	return tx, nil
}

func (txm *TransactionManager) Close(ctx context.Context) error {
	txm.mu.Lock()
	if txm.closed {
		return nil
	}

	txm.closed = true
	txm.mu.Unlock()

	for _, tx := range txm.getActiveTransactions() {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (txm *TransactionManager) getActiveTransactions() []*Transaction {
	txm.mu.RLock()
	defer txm.mu.RUnlock()

	txs := make([]*Transaction, 0, len(txm.activeTransactions))

	for _, tx := range txm.activeTransactions {
		txs = append(txs, tx)
	}

	return txs
}

func (txm *TransactionManager) updateTransaction(tx *Transaction) {
	txm.mu.Lock()
	defer txm.mu.Unlock()

	if tx.done {
		delete(txm.activeTransactions, tx.xid)
	} else {
		txm.activeTransactions[tx.xid] = tx
	}
}

func (txm *TransactionManager) commitTransaction(ctx context.Context, tx *Transaction) error {
	if len(tx.log) == 0 {
		return nil
	}

	ctx = WithTransaction(ctx, nil)

	if err := txm.graph.applyTransaction(ctx, tx); err != nil {
		return err
	}

	last := tx.log[len(tx.log)-1]

	if last.Op != coreapi.JournalOpCommit {
		if err := tx.append(coreapi.JournalEntry{
			Op: coreapi.JournalOpCommit,
		}); err != nil {
			return err
		}
	}

	last = tx.log[len(tx.log)-1]

	return txm.checkpoint.Update(last.Rid, true)
}
