package graphstore

import (
	"sync"
	"sync/atomic"

	"github.com/dgraph-io/sroar"
)

type SerializableBitmap struct {
	*sroar.Bitmap
}

func (b *SerializableBitmap) MarshalBinary() ([]byte, error) {
	return b.ToBufferWithCopy(), nil
}

func (b *SerializableBitmap) UnmarshalBinary(data []byte) error {
	b.Bitmap = sroar.FromBufferWithCopy(data)

	return nil
}

type SerializedBitmapIndex struct {
	LastID   uint64 `json:"lastID"`
	FreeList []byte `json:"freeList"`
	UsedList []byte `json:"usedList"`
}

type SparseBitmapIndex struct {
	mu sync.RWMutex

	idCounter atomic.Uint64
	usedList  *sroar.Bitmap
	freeList  *sroar.Bitmap
}

func NewSparseBitmapIndex() *SparseBitmapIndex {
	return &SparseBitmapIndex{
		freeList: sroar.NewBitmap(),
		usedList: sroar.NewBitmap(),
	}
}

func (b *SparseBitmapIndex) Snapshot() SerializedBitmapIndex {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return SerializedBitmapIndex{
		LastID:   b.idCounter.Load(),
		FreeList: b.freeList.ToBufferWithCopy(),
		UsedList: b.usedList.ToBufferWithCopy(),
	}
}

func (b *SparseBitmapIndex) LoadSnapshot(snap SerializedBitmapIndex) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.idCounter.Store(snap.LastID)
	b.freeList = sroar.FromBufferWithCopy(snap.FreeList)
	b.usedList = sroar.FromBufferWithCopy(snap.UsedList)
}

func (b *SparseBitmapIndex) Allocate() uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	next := b.freeList.NewIterator().Next()

	if next > 0 {
		b.freeList.Remove(next)

		return next
	} else {
		next = b.idCounter.Add(1)
	}

	b.usedList.Set(next)

	return next
}

func (b *SparseBitmapIndex) Free(id uint64) bool {
	if id == 0 {
		panic("id cannot be 0")
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.usedList.Remove(id) {
		return false
	}

	return b.freeList.Set(id)
}

func (b *SparseBitmapIndex) IsUsed(id uint64) bool {
	if id == 0 {
		panic("id cannot be 0")
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.usedList.Contains(id)
}

func (b *SparseBitmapIndex) IsFree(id uint64) bool {
	if id == 0 {
		panic("id cannot be 0")
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.freeList.Contains(id)
}
