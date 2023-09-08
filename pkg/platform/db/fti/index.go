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

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type IndexedDocument struct {
	psi.NodeBase

	Spec ChunkSpec
}

// OnlineIndexQueryHit represents a single search hit in the online index.
type OnlineIndexQueryHit struct {
	Entry    *OnlineIndexEntry
	Distance float32
}

// OnlineIndexEntry represents a single entry in the online index.
// It holds information about the index, chunk, and embedding of a file.
type OnlineIndexEntry struct {
	Index     int64
	Chunk     chunkers.Chunk
	Embedding llm.Embedding
	Document  DocumentReference
}

type OnlineIndex struct {
	Repository *Repository

	m       sync.RWMutex
	idx     faiss.Index
	mapping map[int64]*OnlineIndexEntry
}

// NewOnlineIndex initializes a new OnlineIndex with the given repository.
// It returns a pointer to the created OnlineIndex and an error if any.
//
// NewOnlineIndex takes a repo *Repository as input and creates a new OnlineIndex
// instance. It initializes the OnlineIndex struct with the repo and an empty mapping.
// The function then creates a new Faiss index using faiss.NewIndexFlatIP with a dimension
// of 1536 and assigns it to the idx field of the OnlineIndex struct.
// If an error occurs during the creation of the index, it returns nil and the error.
// Otherwise, it returns a pointer to the created OnlineIndex and nil error.
func NewOnlineIndex(repo *Repository) (*OnlineIndex, error) {
	var err error

	oi := &OnlineIndex{
		Repository: repo,
		mapping:    map[int64]*OnlineIndexEntry{},
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
// For each embedding in the image, the function creates an OnlineIndexEntry, which holds the index, chunk, and embedding of the image.
// It then calls the putEntry() function to store the entry in the repository.
// Finally, it adds the embedding to the faiss index.
// If any error occurs during the process, it returns the error. Otherwise, it returns nil.
func (oi *OnlineIndex) Add(img *ObjectSnapshotImage) error {
	oi.m.Lock()
	defer oi.m.Unlock()

	baseIndex := oi.idx.Ntotal()

	for i, emb := range img.Embeddings {
		entry := &OnlineIndexEntry{
			Index:     baseIndex + int64(i),
			Chunk:     img.Chunks[i],
			Embedding: emb,
			Document:  img.Document,
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

// Query performs a search in the online index using the given query embedding and returns a list of hits.
// Each hit contains the corresponding entry from the index and the distance between the query and the entry embedding.
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

// putEntry stores the given entry in the index with the specified index ID.
//
// The function takes an index ID and an entry as input and stores the entry in the index.
// It first resolves the path to the index directory using the provided repository.
// Then it creates the index directory if it doesn't already exist.
// Next, it resolves the path to store the entry file using the index ID.
// The function then adds the entry to the mapping using the index ID as the key.
// After that, it marshals the entry into JSON format.
// Finally, it writes the JSON data to the specified entry file in the index directory.
// The function returns an error if any error occurs during the process; otherwise, it returns nil.
func (oi *OnlineIndex) putEntry(idx int64, entry *OnlineIndexEntry) error {
	indexPath := oi.Repository.ResolveDbPath("index")

	if err := os.MkdirAll(indexPath, 0755); err != nil {
		return err
	}

	p := oi.Repository.ResolveDbPath("index", strconv.FormatInt(idx, 10))

	oi.mapping[idx] = entry

	data, err := json.Marshal(entry)

	if err != nil {
		return err
	}

	return os.WriteFile(p, data, 0644)
}

// lookupEntry looks up an entry in the online index by its index ID.
// It retrieves the entry from the mapping if it exists, otherwise it reads
// the entry from the index file and adds it to the mapping.
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
