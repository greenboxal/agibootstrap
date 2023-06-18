package fti

import (
	"sync"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"

	"github.com/greenboxal/agibootstrap/pkg/indexing"
)

type IndexObject[K comparable] struct {
	Key        K
	References []int64
}

type FlatKVIndex[K comparable] struct {
	m         sync.RWMutex
	idx       faiss.Index
	entryMap  map[int64]*indexing.IndexEntry[K]
	objectMap map[K]*IndexObject[K]
}

// NewFlatKVIndex initializes a new OnlineIndex with the given repository.
// It returns a pointer to the created FlatKVIndex and an error if any.
//
// NewFlatKVIndex takes a repo *Repository as input and creates a new OnlineIndex
// instance. It initializes the FlatKVIndex struct with the repo and an empty mapping.
// The function then creates a new Faiss index using faiss.NewIndexFlatIP with a dimension
// of 1536 and assigns it to the idx field of the FlatKVIndex struct.
// If an error occurs during the creation of the index, it returns nil and the error.
// Otherwise, it returns a pointer to the created FlatKVIndex and nil error.
func NewFlatKVIndex[K comparable]() (*FlatKVIndex[K], error) {
	var err error

	oi := &FlatKVIndex[K]{
		entryMap:  map[int64]*indexing.IndexEntry[K]{},
		objectMap: map[K]*IndexObject[K]{},
	}

	oi.idx, err = faiss.NewIndexFlatIP(1536)
	if err != nil {
		return nil, err
	}

	return oi, nil
}

// Add adds an image to the online index.
// It takes an ObjectSnapshotImage as input and adds its embeddings to the index.
//
// The function initializes a write lock, which ensures the thread-safety of the online index.
// The function then calculates the base index as the total number of entries in the index.
// For each embedding in the image, the function creates an FlatKVIndexEntry, which holds the index, chunk, and embedding of the image.
// It then calls the putEntry() function to store the entry in the repository.
// Finally, it adds the embedding to the faiss index.
// If any error occurs during the process, it returns the error. Otherwise, it returns nil.
func (oi *FlatKVIndex[K]) Add(key K, value ...llm.Embedding) error {
	oi.m.Lock()
	defer oi.m.Unlock()

	obj := oi.objectMap[key]

	if obj == nil {
		obj = &IndexObject[K]{Key: key}

		oi.objectMap[key] = obj
	}

	for i, emb := range value {
		entry := &indexing.IndexEntry[K]{
			DocumentID: key,
			ChunkIndex: i,
			IndexID:    oi.idx.Ntotal(),
			Embedding:  emb.Embeddings,
			Valid:      true,
		}

		if err := oi.idx.Add(entry.Embedding); err != nil {
			return err
		}

		oi.entryMap[entry.IndexID] = entry
		obj.References = append(obj.References, entry.IndexID)
	}

	return nil
}

func (oi *FlatKVIndex[K]) Delete(key K) bool {
	oi.m.Lock()
	defer oi.m.Unlock()

	obj := oi.objectMap[key]

	if obj == nil {
		return false
	}

	for _, idx := range obj.References {
		entry := oi.entryMap[idx]

		entry.Valid = false
	}

	delete(oi.objectMap, key)

	return true
}

// Query performs a search in the online index using the given query embedding and returns a list of hits.
// Each hit contains the corresponding entry from the index and the distance between the query and the entry embedding.
func (oi *FlatKVIndex[K]) Query(q llm.Embedding, k int64) ([]indexing.SearchHit[K], error) {
	distances, indices, err := oi.idx.Search(q.Embeddings, k)

	if err != nil {
		return nil, err
	}

	hits := make([]indexing.SearchHit[K], len(indices))

	for i, idx := range indices {
		entry := oi.entryMap[idx]

		hits[i] = indexing.SearchHit[K]{
			IndexEntry: *entry,
			Distance:   distances[i],
		}
	}

	return hits, nil
}

// lookupEntry looks up an entry in the online index by its index ID.
// It retrieves the entry from the mapping if it exists, otherwise it reads
// the entry from the index file and adds it to the mapping.
func (oi *FlatKVIndex[K]) lookupEntry(idx int64) (*indexing.IndexEntry[K], error) {
	oi.m.RLock()
	defer oi.m.RUnlock()

	return oi.entryMap[idx], nil
}
