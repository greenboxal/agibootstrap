package coreapi

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type ReplicationManager interface {
	graphfs.ReplicationManager

	CreateConfirmationTracker(ctx context.Context, name string) (ConfirmationTracker, error)
}

type ConfirmationTracker interface {
	Recover() (iterators.Iterator[uint64], error)

	Track(ticket uint64)
	Confirm(ticket uint64)

	Flush() error
	Close() error
}
