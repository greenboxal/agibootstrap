package coreapi

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type ReplicationSlotOptions struct {
	Name       string
	Persistent bool
}

type ReplicationOperations interface {
	CreateReplicationSlot(ctx context.Context, options ReplicationSlotOptions) (ReplicationSlot, error)
}

type ReplicationManager interface {
	ReplicationOperations

	CreateConfirmationTracker(ctx context.Context, name string) (ConfirmationTracker, error)
}

type ConfirmationTracker interface {
	Recover() (iterators.Iterator[uint64], error)

	Track(ticket uint64)
	Confirm(ticket uint64)

	Flush() error
	Close() error
}

type ReplicationMessage struct {
	Xid     uint64
	Entries []*JournalEntry
}

type ReplicationSlot interface {
	Name() string

	GetLastLSN(ctx context.Context) (uint64, error)
	SetLastLSN(ctx context.Context, lsn uint64) error

	Read(ctx context.Context, buffer []ReplicationMessage) (int, error)

	FlushPosition(ctx context.Context) error
}
