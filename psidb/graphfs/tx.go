package graphfs

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
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
		logger:             logging.GetLogger("graphfs/txm"),
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

	it := txm.journal.Iterate(xid, -1)

	recoveredTransactions := map[uint64]*Transaction{}

	for it.Next() {
		var tx *Transaction

		entry := it.Value()

		if entry.Op == JournalOpBegin {
			tx = txm.newTransaction(entry.Xid, true)

			recoveredTransactions[tx.xid] = tx
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

func (txm *TransactionManager) newTransaction(xid uint64, isReplay bool) *Transaction {
	tx := &Transaction{
		xid: xid,
		txm: txm,
	}

	if !isReplay {
		tx.journal = txm.journal
	}

	txm.updateTransaction(tx)

	return tx
}

func (txm *TransactionManager) BeginTransaction() (*Transaction, error) {
	if txm.closed {
		return nil, errors.New("transaction manager already closed")
	}

	xid, err := txm.journal.Write(JournalEntry{
		Op: JournalOpBegin,
	})

	if err != nil {
		return nil, err
	}

	return txm.newTransaction(xid, false), nil
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
	txm.mu.Lock()
	defer txm.mu.Unlock()

	hasBegun := false
	hasFinished := false

	handles := map[int64]NodeHandle{}

	defer func() {
		for _, nh := range handles {
			if err := nh.Close(); err != nil {
				txm.logger.Error(err)
			}
		}
	}()

	for _, entry := range tx.log {
		if hasFinished {
			return errors.New("invalid transaction log")
		}

		if entry.Op != JournalOpBegin && !hasBegun {
			return errors.New("invalid transaction log")
		}

		switch entry.Op {
		case JournalOpBegin:
			hasBegun = true

		case JournalOpCommit:
			hasFinished = true

		case JournalOpRollback:
			hasFinished = true

		case JournalOpCreate:
			nh, err := txm.graph.Open(ctx, *entry.Path, WithOpenNodeCreateIfMissing())

			if err != nil {
				return err
			}

			handles[entry.Inode] = nh

		case JournalOpWrite:
			nh := handles[entry.Inode]

			if nh == nil {
				return errors.New("invalid transaction log")
			}

			if err := nh.Write(ctx, entry.Node); err != nil {
				return err
			}

		case JournalOpSetEdge:
			nh := handles[entry.Inode]

			if nh == nil {
				return errors.New("invalid transaction log")
			}

			if err := nh.SetEdge(ctx, entry.Edge); err != nil {
				return err
			}

		case JournalOpRemoveEdge:
			nh := handles[entry.Inode]

			if nh == nil {
				return errors.New("invalid transaction log")
			}

			if err := nh.RemoveEdge(ctx, entry.Edge.Key); err != nil {
				return err
			}
		}
	}

	return txm.checkpoint.Update(tx.xid)
}

type Transaction struct {
	mu      sync.Mutex
	journal *Journal
	txm     *TransactionManager
	xid     uint64
	log     []*JournalEntry
	done    bool

	dirtyNodes map[int64]*txNode
}

func (tx *Transaction) Append(ctx context.Context, entry JournalEntry) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.done {
		panic(errors.New("transaction already finished"))
	}

	tx.append(entry)

	return nil
}

func (tx *Transaction) append(entry JournalEntry) {
	entry.Xid = tx.xid

	if entry.Ts == 0 {
		entry.Ts = time.Now().UnixNano()
	}

	if tx.journal != nil {
		if _, err := tx.journal.Write(entry); err != nil {
			panic(err)
		}
	}

	tx.log = append(tx.log, &entry)

	switch entry.Op {
	case JournalOpCreate:

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

func (tx *Transaction) Commit(ctx context.Context) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.done {
		return errors.New("transaction already finished")
	}

	tx.done = true

	if err := tx.Append(ctx, JournalEntry{
		Op: JournalOpCommit,
	}); err != nil {
		return err
	}

	tx.txm.updateTransaction(tx)

	return nil
}

func (tx *Transaction) Rollback(ctx context.Context) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.done {
		return errors.New("transaction already finished")
	}

	tx.done = true

	if err := tx.Append(ctx, JournalEntry{
		Op: JournalOpRollback,
	}); err != nil {
		return err
	}

	tx.txm.updateTransaction(tx)

	return nil
}

type txNode struct {
	Inode  int64
	Frozen SerializedNode
	Edges  map[string]SerializedEdge
}

type txNodeHandle struct {
	tx      *Transaction
	inode   *INode
	dentry  *CacheEntry
	closed  bool
	options OpenNodeOptions
}

func (nh *txNodeHandle) Transaction() *Transaction { return nh.tx }
func (nh *txNodeHandle) Inode() *INode             { return nh.inode }
func (nh *txNodeHandle) Entry() *CacheEntry        { return nh.dentry }
func (nh *txNodeHandle) Options() OpenNodeOptions  { return nh.options }

func (nh *txNodeHandle) Read(ctx context.Context) (*SerializedNode, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	n := nh.tx.dirtyNodes[nh.inode.id]

	if n != nil {
		if n.Frozen.Flags&NodeFlagRemoved != 0 {
			return nil, psi.ErrNodeNotFound
		}
	}

	return nh.inode.NodeHandleOperations().Read(ctx, nh)
}

func (nh *txNodeHandle) Write(ctx context.Context, fe *SerializedNode) error {
	if nh.closed {
		return ErrHandleClosed
	}

	return nh.tx.Append(ctx, JournalEntry{
		Op:    JournalOpWrite,
		Inode: nh.inode.id,
		Path:  &fe.Path,
		Node:  fe,
	})
}

func (nh *txNodeHandle) SetEdge(ctx context.Context, edge *SerializedEdge) error {
	if nh.closed {
		return ErrHandleClosed
	}

	edge.Xmax = 0xffffffffffffffff
	edge.Xmin = nh.tx.xid

	return nh.tx.Append(ctx, JournalEntry{
		Op:    JournalOpSetEdge,
		Inode: nh.inode.id,
		Edge:  edge,
	})
}

func (nh *txNodeHandle) RemoveEdge(ctx context.Context, key psi.EdgeKey) error {
	if nh.closed {
		return ErrHandleClosed
	}

	return nh.tx.Append(ctx, JournalEntry{
		Op:    JournalOpRemoveEdge,
		Inode: nh.inode.id,
		Edge:  &SerializedEdge{Key: key},
	})
}

func (nh *txNodeHandle) ReadEdge(ctx context.Context, key psi.EdgeKey) (*SerializedEdge, error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	n := nh.tx.dirtyNodes[nh.inode.id]

	if n != nil {
		e, ok := n.Edges[key.String()]

		if ok {
			if e.Flags&EdgeFlagRemoved != 0 {
				return nil, psi.ErrNodeNotFound
			}

			return &e, nil
		}
	}

	return nh.inode.NodeHandleOperations().ReadEdge(ctx, nh, key)
}

func (nh *txNodeHandle) ReadEdges(ctx context.Context) (iterators.Iterator[*SerializedEdge], error) {
	if nh.closed {
		return nil, ErrHandleClosed
	}

	n := nh.tx.dirtyNodes[nh.inode.id]

	base, err := nh.inode.NodeHandleOperations().ReadEdges(ctx, nh)

	if err != nil && (err == psi.ErrNodeNotFound && n == nil) {
		return nil, err
	}

	if base == nil {
		base = iterators.Empty[*SerializedEdge]()
	}

	if n == nil {
		return base, nil
	}

	dirtyEdges := iterators.FromMap(n.Edges)
	seenMap := map[string]struct{}{}

	return iterators.NewIterator(func() (*SerializedEdge, bool) {
		for {
			if base != nil {
				if !base.Next() {
					base = nil
					continue
				}

				e := base.Value()

				if e2, ok := n.Edges[e.Key.String()]; ok {
					e = &e2
					seenMap[e.Key.String()] = struct{}{}
				}

				if e.Flags&EdgeFlagRemoved != 0 {
					continue
				}

				return e, true
			} else {
				if !dirtyEdges.Next() {
					return nil, false
				}

				e := dirtyEdges.Value()

				if _, ok := seenMap[e.K]; ok {
					continue
				}

				if e.V.Flags&EdgeFlagRemoved != 0 {
					continue
				}

				return &e.V, true
			}
		}
	}), nil
}

func (nh *txNodeHandle) Close() error {
	if nh.closed {
		return nil
	}

	if nh.dentry != nil {
		nh.dentry.Unref()
		nh.dentry = nil
	}

	if nh.inode != nil {
		nh.inode.Unref()
		nh.inode = nil
	}

	nh.closed = true

	return nil
}

var ctxKeyTransaction = &struct{}{}

func WithTransaction(ctx context.Context, tx *Transaction) context.Context {
	return context.WithValue(ctx, ctxKeyTransaction, tx)
}

func GetTransaction(ctx context.Context) *Transaction {
	tx, _ := ctx.Value(ctxKeyTransaction).(*Transaction)

	return tx
}
