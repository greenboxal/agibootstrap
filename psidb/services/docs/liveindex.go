package docs

import (
	"os"
	"path"
	"sync"

	"github.com/DataIntelligenceCrew/go-faiss"
)

type LiveIndex struct {
	uuid string

	workspaceDir string
	snapshotPath string

	mu    sync.RWMutex
	im    *IndexManager
	faiss faiss.Index
}

func NewLiveIndex(im *IndexManager, uuid string, workspaceDir string) (*LiveIndex, error) {
	li := &LiveIndex{
		im:           im,
		uuid:         uuid,
		workspaceDir: workspaceDir,
	}

	if err := os.MkdirAll(li.workspaceDir, 0755); err != nil {
		return nil, err
	}

	li.snapshotPath = path.Join(workspaceDir, "index.faiss")

	if _, err := os.Stat(li.snapshotPath); err == nil {
		idx, err := faiss.ReadIndex(li.snapshotPath, faiss.IOFlagMmap)

		if err != nil {
			return nil, err
		}

		li.faiss = idx
	} else {
		idx, err := faiss.IndexFactory(1536, "IDMap,Flat", faiss.MetricInnerProduct)

		if err != nil {
			return nil, err
		}

		li.faiss = idx

		if err := li.Flush(); err != nil {
			return nil, err
		}
	}

	return li, nil
}

func (li *LiveIndex) AddChunk(chunk *IndexEntryChunk) error {
	li.mu.Lock()
	defer li.mu.Unlock()

	if err := li.faiss.AddWithIDs(chunk.Embeddings, []int64{chunk.Ordinal}); err != nil {
		return err
	}

	return li.flushUnlocked()
}

func (li *LiveIndex) RemoveChunk(ord int64) error {
	li.mu.Lock()
	defer li.mu.Unlock()

	sel, err := faiss.NewIDSelectorRange(ord, ord+1)

	if err != nil {
		return err
	}

	defer sel.Delete()

	n, err := li.faiss.RemoveIDs(sel)

	if err != nil {
		return err
	}

	if n == 0 {
		return nil
	}

	return li.flushUnlocked()
}

func (li *LiveIndex) QueryChunks(query *IndexEntryChunk, limit int) ([]float32, []int64, error) {
	li.mu.RLock()
	defer li.mu.RUnlock()

	return li.faiss.Search(query.Embeddings, int64(limit))
}

func (li *LiveIndex) Flush() error {
	li.mu.RLock()
	defer li.mu.RUnlock()

	return li.flushUnlocked()
}

func (li *LiveIndex) flushUnlocked() error {
	if err := faiss.WriteIndex(li.faiss, li.snapshotPath); err != nil {
		return err
	}

	return nil
}

func (li *LiveIndex) Close() error {
	li.mu.Lock()
	defer li.mu.Unlock()

	if err := li.flushUnlocked(); err != nil {
		return err
	}

	li.faiss.Delete()
	li.faiss = nil

	li.im.notifyClose(li)

	return nil
}
