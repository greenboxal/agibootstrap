package fti

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
	"sync"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
)

// OnlineIndexQueryHit represents a single search hit in the online index.
type OnlineIndexQueryHit struct {
	Entry    *OnlineIndexEntry
	Distance float32
}

// TODO: Write documentation for this file
type OnlineIndexEntry struct {
	Index     int64
	Chunk     chunkers.Chunk
	Embedding llm.Embedding
}

// TODO: Write documentation for this file
type OnlineIndex struct {
	Repository *Repository

	m       sync.RWMutex
	idx     faiss.Index
	mapping map[int64]*OnlineIndexEntry
}

// TODO: Write documentation for this file
func NewOnlineIndex(repo *Repository) (*OnlineIndex, error) {
	var err error

	oi := &OnlineIndex{
		Repository: repo,

		mapping: map[int64]*OnlineIndexEntry{},
	}

	oi.idx, err = faiss.NewIndexFlatIP(1536)

	if err != nil {
		return nil, err
	}

	return oi, nil
}

// TODO: Write documentation for this file
func (oi *OnlineIndex) Add(img *ObjectSnapshotImage) error {
	oi.m.Lock()
	defer oi.m.Unlock()

	baseIndex := oi.idx.Ntotal()

	for i, emb := range img.Embeddings {
		entry := &OnlineIndexEntry{
			Index:     baseIndex + int64(i),
			Chunk:     img.Chunks[i],
			Embedding: emb,
		}

		if err := oi.putEntry(entry.Index, entry); err != nil {
			return err
		}

		if err := oi.idx.Add(emb.Embeddings); err != nil {
			return err
		}
	}

	return nil
}

// TODO: Write documentation for this file
func (oi *OnlineIndex) Query(q llm.Embedding, k int64) ([]OnlineIndexQueryHit, error) {
	distances, indices, err := oi.idx.Search(q.Embeddings, k)

	if err != nil {
		return nil, err
	}

	hits := make([]OnlineIndexQueryHit, len(indices))

	for i, idx := range indices {
		entry, err := oi.lookupEntry(idx)

		if err != nil {
			return nil, err
		}

		hits[i] = OnlineIndexQueryHit{
			Entry:    entry,
			Distance: distances[i],
		}
	}

	return hits, nil
}

// TODO: Write documentation for this file
func (oi *OnlineIndex) putEntry(idx int64, entry *OnlineIndexEntry) error {
	indexPath := oi.Repository.ResolveDbPath("index")

	if err := os.MkdirAll(indexPath, os.ModePerm); err != nil {
		return err
	}

	p := oi.Repository.ResolveDbPath("index", strconv.FormatInt(idx, 10))

	oi.mapping[idx] = entry

	data, err := json.Marshal(entry)

	if err != nil {
		return err
	}

	return os.WriteFile(p, data, os.ModePerm)
}

// TODO: Write documentation for this file
func (oi *OnlineIndex) lookupEntry(idx int64) (*OnlineIndexEntry, error) {
	oi.m.Lock()
	defer oi.m.Unlock()

	existing := oi.mapping[idx]

	if existing == nil {
		indexFilePath := oi.Repository.ResolveDbPath("index", strconv.FormatInt(idx, 10))

		data, err := ioutil.ReadFile(indexFilePath)
		if err != nil {
			return nil, err
		}

		entry := &OnlineIndexEntry{}
		err = json.Unmarshal(data, entry)
		if err != nil {
			return nil, err
		}

		oi.mapping[idx] = entry
		existing = entry
	}

	return existing, nil
}
