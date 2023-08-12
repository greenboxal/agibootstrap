package graphindex

import (
	"context"
	"os"
	"path"
	"sync"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type faissIndex struct {
	mu sync.RWMutex

	id   string
	path string
	d    int

	m     *Manager
	ds    datastore.Batching
	index faiss.Index

	isDirty     bool
	saveOnClose bool
}

func newFaissIndex(m *Manager, stateDir string, id string, d int) (*faissIndex, error) {
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, err
	}

	idx, err := faiss.NewIndexFlatIP(d)

	if err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions
	ds, err := badger.NewDatastore(stateDir, &opts)

	if err != nil {
		return nil, err
	}

	return &faissIndex{
		id:          id,
		m:           m,
		d:           d,
		index:       idx,
		ds:          ds,
		path:        stateDir,
		saveOnClose: true,
	}, nil
}

func (f *faissIndex) Dimensions() int { return f.d }

func (f *faissIndex) IndexNode(ctx context.Context, req IndexNodeRequest) (IndexedItem, error) {
	item := IndexedItem{
		ChunkIndex: req.ChunkIndex,
		ChunkLink:  req.ChunkLink,
		Path:       req.Path,
		Link:       req.Link,
		Embeddings: req.Embeddings,
	}

	x := req.Embeddings.ToFloat32Slice(nil)

	f.mu.Lock()
	defer f.mu.Unlock()

	existingIndex, err := f.retrieveItemIndex(ctx, item)

	if err != nil {
		if errors.Is(err, psi.ErrNodeNotFound) {
			existingIndex = 0xFFFFFFFFFFFFFFFF
		} else {
			return IndexedItem{}, err
		}
	}

	if existingIndex != 0xFFFFFFFFFFFFFFFF {
		item.Index = int64(existingIndex)
	} else {
		item.Index = f.index.Ntotal()

		if err := f.index.Add(x); err != nil {
			return IndexedItem{}, errors.Wrap(err, "failed to add to index")
		}
	}

	f.isDirty = true

	if err := f.storeItem(ctx, item); err != nil {
		return IndexedItem{}, err
	}

	return item, nil
}

func (f *faissIndex) Search(ctx context.Context, req SearchRequest) (iterators.Iterator[BasicSearchHit], error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	x := req.Query.ToFloat32Slice(nil)
	distances, ids, err := f.index.Search(x, int64(req.Limit))

	if err != nil {
		return nil, err
	}

	hits := make([]BasicSearchHit, len(ids))

	for i, id := range ids {
		if id == -1 {
			continue
		}

		item, err := f.retrieveItem(ctx, id)

		if err != nil {
			return nil, err
		}

		hits[i] = BasicSearchHit{
			IndexedItem: item,
			Score:       distances[i],
		}
	}

	return iterators.FromSlice(hits), nil
}

func (f *faissIndex) Rebuild(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.index.Reset(); err != nil {
		return err
	}

	it, err := psids.List(ctx, f.ds, dsKeyIndexItemPrefix(f.id))

	if err != nil {
		return err
	}

	for it.Next() {
		if err := f.index.Add(it.Value().Embeddings.ToFloat32Slice(nil)); err != nil {
			return err
		}
	}

	return nil
}

func (f *faissIndex) Truncate(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.index == nil {
		return nil
	}

	if err := f.index.Reset(); err != nil {
		return err
	}

	f.isDirty = true

	return nil
}

func (f *faissIndex) Save() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.isDirty {
		return nil
	}

	if err := f.saveSnapshot(f.getSnapshotPath()); err != nil {
		return err
	}

	f.isDirty = false

	return nil
}

func (f *faissIndex) Load() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	snapshotPath := f.getSnapshotPath()

	if _, err := os.Stat(snapshotPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		return f.restoreSnapshot(snapshotPath)
	}

	return f.createNew()
}

func (f *faissIndex) Close() error {
	if f.index != nil {
		if f.saveOnClose {
			if err := f.Save(); err != nil {
				return nil
			}
		}

		//f.index.Delete()
		//f.index = nil
	}

	/*if f.ds != nil {
		if err := f.ds.Close(); err != nil {
			return err
		}
	}*/

	return nil
}

func (f *faissIndex) createNew() error {
	idx, err := faiss.NewIndexFlatIP(f.d)

	if err != nil {
		return err
	}

	f.index = idx

	return nil
}

func (f *faissIndex) restoreSnapshot(path string) error {
	if f.index != nil {
		f.index.Delete()
	}

	idx, err := faiss.ReadIndex(path, faiss.IOFlagMmap)

	if err != nil {
		return err
	}

	f.index = idx
	f.isDirty = true

	return nil
}

func (f *faissIndex) saveSnapshot(path string) error {
	if f.index == nil {
		return errors.New("index not loaded")
	}

	return faiss.WriteIndex(f.index, path)
}

func (f *faissIndex) getSnapshotPath() string {
	return path.Join(f.path, "faiss.bin")
}

func (f *faissIndex) storeItem(ctx context.Context, item IndexedItem) error {
	batch, err := f.ds.Batch(ctx)

	if err != nil {
		return err
	}

	if err := psids.Put(ctx, f.ds, dsKeyIndexItem(f.id, item.Index), item); err != nil {
		return err
	}

	if err := psids.Put(ctx, f.ds, dsKeyInvertedIndex(item.Identity()), uint64(item.Index)); err != nil {
		return err
	}

	return batch.Commit(ctx)
}

func (f *faissIndex) retrieveItem(ctx context.Context, id int64) (IndexedItem, error) {
	return psids.Get(ctx, f.ds, dsKeyIndexItem(f.id, id))
}

func (f *faissIndex) retrieveItemIndex(ctx context.Context, item IndexedItem) (uint64, error) {
	return psids.Get(ctx, f.ds, dsKeyInvertedIndex(f.id, item.Identity()))
}
