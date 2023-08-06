package graphfs

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
)

type TransactionManager struct {
	logger     *zap.SugaredLogger
	journal    *Journal
	checkpoint Checkpoint
	graph      *VirtualGraph

	mu                 sync.RWMutex
	activeTransactions map[uint64]*Transaction

	closed bool
}

func NewTransactionManager(graph *VirtualGraph, journal *Journal, checkpoint Checkpoint) *TransactionManager {
	return &TransactionManager{
		logger:             logging.GetLogger("graphfs/transactionManager"),
		graph:              graph,
		journal:            journal,
		checkpoint:         checkpoint,
		activeTransactions: make(map[uint64]*Transaction),
	}
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

	recoveredTransactions := map[uint64]*Transaction{}

	for it.Next() {
		var tx *Transaction

		entry := it.Value()

		if entry.Op == JournalOpBegin {
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

		if entry.Op == JournalOpCommit {
			if err := txm.commitTransaction(ctx, tx); err != nil {
				return err
			}
		} else if entry.Op == JournalOpRollback {
			if err := tx.Rollback(ctx); err != nil {
				return err
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
		txm: txm,

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

	if last.Op != JournalOpCommit {
		if err := tx.append(JournalEntry{
			Op: JournalOpCommit,
		}); err != nil {
			return err
		}
	}

	last = tx.log[len(tx.log)-1]

	return txm.checkpoint.Update(last.Rid)
}
