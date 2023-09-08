package graphfs

import (
	"context"
	"io"
	"sync"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type replicationManager struct {
	mu sync.Mutex

	vg    *VirtualGraph
	slots map[string]*replicationSlot
}

func newReplicationManager(vg *VirtualGraph) *replicationManager {
	return &replicationManager{
		vg:    vg,
		slots: map[string]*replicationSlot{},
	}
}

func (r *replicationManager) CreateReplicationSlot(ctx context.Context, options coreapi.ReplicationSlotOptions) (coreapi.ReplicationSlot, error) {
	if options.Name == "" {
		return nil, errors.New("name must not be empty")
	}

	if !options.Persistent {
		return newReplicationSlot(r.vg, r.vg.transactionManager.journal, options), nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if slot := r.slots[options.Name]; slot != nil && !slot.closed {
		return slot, nil
	}

	slot := newReplicationSlot(r.vg, r.vg.transactionManager.journal, options)

	if err := slot.ensureLoaded(ctx); err != nil {
		return nil, err
	}

	r.slots[options.Name] = slot

	return slot, nil
}

func (r *replicationManager) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, slot := range r.slots {
		if err := slot.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

type replicationSlot struct {
	mu sync.RWMutex

	vg      *VirtualGraph
	journal coreapi.Journal
	options coreapi.ReplicationSlotOptions

	lastLsn uint64

	loaded bool
	closed bool

	recoveredTransactions map[uint64]*Transaction
}

func newReplicationSlot(vg *VirtualGraph, journal coreapi.Journal, options coreapi.ReplicationSlotOptions) *replicationSlot {
	return &replicationSlot{
		vg:      vg,
		journal: journal,
		options: options,

		recoveredTransactions: map[uint64]*Transaction{},
	}
}

func (r *replicationSlot) Name() string { return r.options.Name }

func (r *replicationSlot) GetLastLSN(ctx context.Context) (uint64, error) {
	if err := r.ensureLoaded(ctx); err != nil {
		return 0, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.lastLsn, nil
}

func (r *replicationSlot) SetLastLSN(ctx context.Context, lsn uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return io.EOF
	}

	r.lastLsn = lsn

	return nil
}

func (r *replicationSlot) Read(ctx context.Context, buffer []coreapi.ReplicationMessage) (i int, err error) {
	if err := r.ensureLoaded(ctx); err != nil {
		return 0, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return 0, io.EOF
	}

	for i < len(buffer) {
		var tx *Transaction

		entry := &coreapi.JournalEntry{}

		if _, err := r.journal.Read(r.lastLsn, entry); err != nil {
			if err == io.EOF {
				break
			}

			return i, err
		}

		r.lastLsn = r.lastLsn + 1

		if entry.Op == coreapi.JournalOpBegin {
			tx = &Transaction{
				xid:        entry.Xid,
				dirtyNodes: make(map[int64]*txNode),
			}

			r.recoveredTransactions[entry.Xid] = tx
		} else {
			tx = r.recoveredTransactions[entry.Xid]
		}

		if tx == nil {
			return i, errors.New("invalid transaction id")
		}

		if tx.done {
			continue
		}

		tx.log = append(tx.log, entry)

		if entry.Op == coreapi.JournalOpCommit {
			buffer[i] = coreapi.ReplicationMessage{
				Xid:     entry.Xid,
				Entries: tx.log,
			}

			i++

			tx.done = true
		} else if entry.Op == coreapi.JournalOpRollback {
			tx.done = true
		}

		if tx.done {
			delete(r.recoveredTransactions, entry.Xid)
		}
	}

	err = nil

	return
}

func (r *replicationSlot) FlushPosition(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.flushPosition(ctx)
}

func (r *replicationSlot) Close(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	if err := r.flushPosition(ctx); err != nil {
		return err
	}

	return nil
}

func (r *replicationSlot) flushPosition(ctx context.Context) error {
	if !r.options.Persistent {
		return nil
	}

	return psids.Put(ctx, r.vg.ds, dsKeyReplicationSlotLSN(r.options.Name), r.lastLsn)
}

func (r *replicationSlot) ensureLoaded(ctx context.Context) error {
	if r.closed {
		return io.EOF
	}

	if r.loaded {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return io.EOF
	}

	if r.loaded {
		return nil
	}

	lsn, err := psids.Get(ctx, r.vg.ds, dsKeyReplicationSlotLSN(r.options.Name))

	if errors.Is(err, psi.ErrNodeNotFound) {
		lsn = 1
	} else if err != nil {
		return err
	}

	r.lastLsn = lsn
	r.loaded = true

	return nil
}

var dsKeyReplicationSlotLSN = psids.KeyTemplate[uint64]("replication-slot-lsn:%s")
