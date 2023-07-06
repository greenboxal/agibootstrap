package indextree

import (
	"context"
	"sync"

	"github.com/DataIntelligenceCrew/go-faiss"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Index struct {
	psi.NodeBase

	mu      sync.RWMutex
	backend *faiss.IndexFlat
}

var IndexType = psi.DefineNodeType[*Index]()

func NewIndex() (*Index, error) {
	i := &Index{}

	if err := i.Init(); err != nil {
		return nil, err
	}

	return i, nil
}

func (i *Index) Init() error {
	var err error

	i.NodeBase.Init(i, psi.WithNodeType(IndexType))

	i.backend, err = faiss.NewIndexFlatIP(1536)

	if err != nil {
		return err
	}

	collectionsfx.ObserveList(i.ChildrenList(), func(ev collectionsfx.ListChangeEvent[psi.Node]) {
		for ev.Next() {
			if ev.WasAdded() {
				for _, v := range ev.AddedSlice() {
					if v, ok := v.(*IndexEntry); ok {
						i.addIndexEntry(v)
					}
				}
			} else if ev.WasRemoved() {
				for _, v := range ev.RemovedSlice() {
					if v, ok := v.(*IndexEntry); ok {
						i.removeIndexEntry(v)
					}
				}
			}
		}
	})

	return nil
}

func (i *Index) addIndexEntry(v *IndexEntry) {
	if v.Valid {
		return
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	v.Index = i.backend.Ntotal()
	v.Valid = true

	v.Invalidate()

	if err := i.backend.Add(v.Embedding); err != nil {
		panic(err)
	}

	i.InsertChildrenAt(int(v.Index), v)
}

func (i *Index) removeIndexEntry(v *IndexEntry) {
	if !v.Valid {
		return
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	if v.Index < 0 {
		return
	}

	sel, err := faiss.NewIDSelectorBatch([]int64{v.Index})

	if err != nil {
		panic(err)
	}

	defer sel.Delete()

	if _, err = i.backend.RemoveIDs(sel); err != nil {
		panic(err)
	}

	v.Valid = false
	v.Index = -1
}

func (i *Index) OnUpdate(ctx context.Context) error {
	if err := i.NodeBase.OnUpdate(ctx); err != nil {
		return err
	}

	return nil
}

func (i *Index) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.backend != nil {
		i.backend.Delete()
		i.backend = nil
	}

	return nil
}

type IndexEntry struct {
	psi.NodeBase

	Valid   bool  `json:"valid"`
	Index   int64 `json:"index"`
	ChunkID int64 `json:"chunk_id"`

	Embedding []float32 `json:"embedding"`
}

var IndexEntryType = psi.DefineNodeType[*IndexEntry]()

func NewIndexEntry() *IndexEntry {
	i := &IndexEntry{}

	i.Init(i, psi.WithNodeType(IndexEntryType))

	return i
}
