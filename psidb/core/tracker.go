package core

import (
	"context"
	"errors"
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/ipfs/go-datastore"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type ConfirmationTracker struct {
	ds   coreapi.DataStore
	name string
	key  datastore.Key

	mu    sync.RWMutex
	bmp   *roaring64.Bitmap
	dirty bool
}

func NewConfirmationTracker(ctx context.Context, ds coreapi.DataStore, name string) (*ConfirmationTracker, error) {
	ct := &ConfirmationTracker{
		ds:   ds,
		name: name,
		key:  datastore.NewKey("/_ctracker/" + name),
		bmp:  roaring64.New(),
	}

	data, err := ds.Get(ctx, ct.key)

	if err != nil && !errors.Is(err, datastore.ErrNotFound) {
		return nil, err
	} else if err == nil {
		if err := ct.bmp.UnmarshalBinary(data); err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}
	}

	return ct, nil
}

func (c *ConfirmationTracker) Recover() (iterators.Iterator[uint64], error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	it := c.bmp.Iterator()

	return iterators.NewIterator(func() (uint64, bool) {
		if !it.HasNext() {
			return 0, false
		}

		return it.Next(), true
	}), nil
}

func (c *ConfirmationTracker) Track(ticket uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.bmp.CheckedAdd(ticket) {
		c.dirty = true
	}
}

func (c *ConfirmationTracker) Confirm(ticket uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.bmp.CheckedRemove(ticket) {
		c.dirty = true
	}
}

func (c *ConfirmationTracker) Flush() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.dirty {
		return nil
	}

	data, err := c.bmp.ToBytes()

	if err != nil {
		return err
	}

	if err := c.ds.Put(context.Background(), c.key, data); err != nil {
		return err
	}

	c.dirty = false

	return nil
}

func (c *ConfirmationTracker) Close() error {
	return c.Flush()
}
