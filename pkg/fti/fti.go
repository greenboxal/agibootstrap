package fti

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/fs"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"github.com/pkg/errors"
	ignore "github.com/sabhiram/go-gitignore"
)

/*
# File Tree Index Design Document

This technical document outlines the design and structure of a File Tree Index (FTI), a pivotal component of our system designed for indexing file-based repositories using embedding vectors. The FTI provides a spatially efficient, content-addressable mapping of files within a repository to unique embedding vectors, obtained via an external API, such as OpenAI's Embeddings API.

## Overview of File Tree Indexing

The FTI is an innovative indexing approach that dissects each file into manageable 'chunks', assigns unique embedding vectors to these chunks, and organizes them in a way that facilitates efficient tracking, retrieval, and storage of file data.

## FTI Structure

The FTI is situated within a .fti directory at the root of the repository, structured as follows:

```
/path/to/repository/.fti/
    - objects/
          - <object hash>/
              - <chunkSize>m<overlap>.bin
```

### Object Entities

In this context, an object refers to a versioned snapshot of specific content, organized in a content-addressable fashion. These objects are cataloged in an 'objects' subdirectory under the .fti directory. Each object is stored in a separate directory, labeled with the hash of the content it represents.

For every distinct chunking configuration (specified by chunk size and overlap), a unique object snapshot file is created and named following the pattern `<chunkSize>m<overlap>.bin`.

The object snapshot files are represented as 2D square images, with each image encapsulating the embeddings of its respective chunks. The embeddings are displayed as RGB squares arranged within each chunk. This visualization offers a graphical interpretation of the distribution and structure of the object snapshot.

### Object Snapshot File Layout

The object snapshot file is a binary file following a specific byte layout to effectively represent the embeddings of each file chunk. Each file is named in the `<chunkSize>m<overlap>.bin` format and located in the directory named after the hash of the object it represents.

The file begins with a header that includes the object's hash, followed by the chunk size and overlap in bytes. Following the header, each embedding is represented as a sequence of bytes for each RGB square, with each color component represented as a byte (8 bits).

Given that each embedding vector is visualized as an RGB square, the byte sequence for a single embedding would be in the format:

```
R1 G1 B1 R2 G2 B2 ... Rn Gn Bn
```

Where `Rn`, `Gn`, `Bn` are the byte representations of the RGB components of the nth square in the 2D image.

### Snapshot Metadata

An integral part of the FTI repository is the Snapshot Metadata, a JSON encoded file which provides descriptive or ancillary data about each object snapshot. The metadata is stored in the .fti directory under the path: `.fti/objects/<object hash>/metadata.json`.

The Snapshot Metadata includes crucial information about the object snapshot such as:

- The hash of the object snapshot (which corresponds to the object directory name)
- The chunking configuration (chunk size and overlap)
- The date and time the snapshot was created
- The size of the snapshot file
- The total number of chunks in the snapshot

An example of the Snapshot Metadata is:

```json
{
    "objectHash": "<object hash>",
    "chunkSize": "<chunk size>",
    "overlap": "<overlap>",
    "creationTime": "<creation timestamp>",
    "fileSize": "<snapshot file size>",
    "totalChunks": "<total number of chunks>"
}
```

This metadata plays a crucial role in retrieving, managing, and understanding the indexed data in the FTI. It provides an easy way to perform operations on

 the object snapshot without having to load or read the actual snapshot file.

### Inverse Index

An additional component of the FTI is the Inverse Index, a data structure stored under `.fti/index/`. The Inverse Index is a binary tree structure optimized for swift and efficient in-memory retrieval, designed to facilitate direct mapping from int64 identifiers to a tuple that comprises the object hash and the chunk index, expressed as `int64 -> (object hash, chunk index)`.

```
/path/to/repository/.fti/index/
     - inverseIndex.bin
```

In this directory structure, `inverseIndex.bin` is the binary file representing the Inverse Index binary tree. The optimized binary tree structure enables rapid data lookups, improved storage efficiency, and fast data retrieval, making it an essential component for achieving high-performance levels in large-scale systems.

### Inverse Index File Layout

The Inverse Index file, `inverseIndex.bin`, is structured as a binary tree for fast and efficient retrieval of data. It maps an int64 identifier to a tuple consisting of an object hash and a chunk index.

The binary file layout begins with a header specifying the total number of entries in the binary tree. Each subsequent entry comprises the int64 identifier (8 bytes), followed by the object hash (typically SHA-256, thus 32 bytes), and then the chunk index (as an int64, thus another 8 bytes).

The structure for a single entry would be in the format:

```
<int64 identifier> <object hash> <chunk index>
```

To ensure consistency and efficient retrieval, entries within the `inverseIndex.bin` file are sorted by the int64 identifier, adhering to the inherent sorted nature of a binary search tree. The layout design facilitates the in-memory loading of the tree, optimizing the retrieval time and reducing the IO overhead.

### Repository Configuration

The repository configuration for the FTI is stored in a JSON file located at `.fti/config.json`. This file serves as the central source for the configuration parameters governing the operation of the FTI within the repository.

The repository configuration file contains several key parameters:

- **ChunkSizes**: An array of integers defining the sizes of the chunks to be used for breaking down files in the repository.
- **Overlaps**: An array of integers specifying the overlaps between chunks, for the file chunking process.
- **EmbeddingAPI**: A string that specifies the external API to be used for generating embeddings (e.g., "OpenAI").
- **EmbeddingDimensions**: An integer representing the dimension of the embeddings created by the Embedding API.

A sample `config.json` might look like this:

```json
{
    "ChunkSizes": [1024, 2048, 4096],
    "Overlaps": [256, 512, 1024],
    "EmbeddingAPI": "OpenAI",
    "EmbeddingDimensions": 768
}
```

This configuration file provides a flexible and straightforward mechanism to tweak the operation of the FTI according to the requirements of the specific repository. It ensures that the FTI can be easily adapted and optimized for various types of data and computational constraints.

### Initialization Operation

The initialization operation, often referred to as the 'init' operation, is a vital process that sets up the necessary structure and configuration for the FTI within a repository. The operation is typically performed only once at the beginning or when the existing FTI needs to be reset.

The init operation primarily involves the creation of the `.fti` directory at the root of the repository and the establishment of its subdirectories (`objects/` and `index/`).

Simultaneously, the `config.json` file is created in the `.fti` directory, with the configuration parameters defined by the user or set to some default values if not provided. The configuration file contains several parameters that guide the operation of the FTI, as described in the previous sections.

Once the configuration file is established, the FTI system is ready to index files in the repository. It should be noted that during the init operation, no files are indexed. The process only sets up the necessary infrastructure for subsequent indexing.

Here is a pseudo-code for the init operation:

```go
function initFTI(RepoPath string, Config FTIConfig) {
    // Create .fti directory at the root of the repository
    createDirectory(RepoPath + "/.fti")

    // Create 'objects' and 'index' subdirectories
    createDirectory(RepoPath + "/.fti/objects")
    createDirectory(RepoPath + "/.fti/index")

    // Create the configuration file
    writeToFile(RepoPath + "/.fti/config.json", Config)
}
```

It's important to remember that initialization is a critical operation that lays the groundwork for the FTI. Care should be taken to ensure that all parameters are correctly set, as they play a significant role in the performance and efficiency of the indexing process.

### Update Operation

The update operation is the key process that indexes files within the repository, creating object snapshots, and updating the FTI. It is designed to be run either manually whenever there are significant changes to the repository or set up to execute automatically at regular intervals.

The update operation proceeds in several steps:

1. **File Scanning**: The entire repository is scanned to identify new files, modified files, and deleted files since the last update.

2. **Chunking**: The new and modified files are divided into chunks according to the chunk sizes specified in the configuration file.

3. **Embedding Generation**: For each chunk, an embedding is generated using the API specified in the configuration file.

4. **Object Snapshot Creation**: The generated embeddings are compiled into object snapshots, which are then stored in the appropriate directories under the `.fti/objects` path.

5. **Metadata Generation**: Metadata for each object snapshot is generated and saved in the `metadata.json` file in the corresponding object directory.

6. **Inverse Index Update**: The Inverse Index is updated to include references to the new and modified files and to remove entries corresponding to deleted files.

Here is a simplified pseudo-code for the update operation:

```go
function updateFTI(RepoPath string) {
    // Load the configuration file
    Config = loadConfig(RepoPath + "/.fti/config.json")

    // Scan the repository for changes
    NewFiles, ModifiedFiles, DeletedFiles = scanRepository(RepoPath)

    // Process new and modified files
    for each file in NewFiles + ModifiedFiles {
        // Generate chunks for the file
        Chunks = chunkFile(file, Config.ChunkSizes, Config.Overlaps)

        // Generate embeddings for each chunk
        Embeddings = generateEmbeddings(Chunks, Config.EmbeddingAPI)

        // Create object snapshot and metadata
        createObjectSnapshot(Embeddings)
        createMetadata(file)

        // Update the Inverse Index
        updateInverseIndex(file)
    }

    // Process deleted files
    for each file in DeletedFiles {
        // Update the Inverse Index
        removeEntryFromInverseIndex(file)
    }
}
```

This update operation ensures that the FTI stays current with the repository content, reflecting the latest changes and updates. It is a key component of maintaining an efficient, accurate index of the repository content.

### Query Operation

The query operation allows users to retrieve information about specific files in the repository based on their embeddings. This operation is fundamental to the value provided by the FTI, enabling users to execute complex queries and extract valuable insights from the indexed data.

The query operation follows several steps:

1. **Embedding Generation**: The query input, often a chunk of data similar to those indexed, is transformed into an embedding using the same API specified in the configuration file.

2. **Inverse Index Lookup**: The generated embedding is used to look up the corresponding file in the Inverse Index. The lookup retrieves the object hash and the chunk index.

3. **Object Retrieval**: Using the retrieved object hash, the relevant object snapshot is located within the `.fti/objects/` directory.

4. **Metadata Extraction**: Metadata for the object snapshot is loaded from the `metadata.json` file in the object directory. This metadata provides additional context and details about the object.

5. **Chunk Retrieval**: Within the object snapshot, the chunk referenced by the chunk index is extracted.

The exact mechanism for comparing the query embedding with the embeddings in the Inverse Index may vary depending on the specifics of the Embedding API and the dimensions of the embeddings.

Here is a pseudo-code representation of the query operation:

```go
function queryFTI(QueryInput string, RepoPath string) {
    // Load the configuration file
    Config = loadConfig(RepoPath + "/.fti/config.json")

    // Generate an embedding for the query input
    QueryEmbedding = generateEmbedding(QueryInput, Config.EmbeddingAPI)

    // Perform a lookup in the Inverse Index
    ObjectHash, ChunkIndex = lookupInverseIndex(QueryEmbedding)

    // Retrieve the object snapshot and metadata
    ObjectSnapshot = retrieveObjectSnapshot(ObjectHash)
    Metadata = loadMetadata(ObjectHash)

    // Extract the relevant chunk
    Chunk = extractChunk(ObjectSnapshot, ChunkIndex)

    // Return the retrieved information
    return Chunk, Metadata
}
```

The query operation offers a powerful tool for users to delve into their indexed data, providing quick and easy access to specific file chunks based on their content. It is an essential part of the FTI's functionality, allowing the system to provide the content-addressable file indexing that is its primary purpose.
*/

// Reference types:
//type llm.Embedder interface {
//	MaxTokensPerChunk() int
//
//	GetEmbeddings(ctx context.Context, chunks []string) ([]Embedding, error)
//}
//
//type llm.Embedding struct {
//	Embeddings []float32
//}
//
//type faiss.Index interface {
//	// D returns the dimension of the indexed vectors.
//	D() int
//
//	// IsTrained returns true if the index has been trained or does not require
//	// training.
//	IsTrained() bool
//
//	// Ntotal returns the number of indexed vectors.
//	Ntotal() int64
//
//	// MetricType returns the metric type of the index.
//	MetricType() int
//
//	// Train trains the index on a representative set of vectors.
//	Train(x []float32) error
//
//	// Add adds vectors to the index.
//	Add(x []float32) error
//
//	// AddWithIDs is like Add, but stores xids instead of sequential IDs.
//	AddWithIDs(x []float32, xids []int64) error
//
//	// Search queries the index with the vectors in x.
//	// Returns the IDs of the k nearest neighbors for each query vector and the
//	// corresponding distances.
//	Search(x []float32, k int64) (distances []float32, labels []int64, err error)
//
//	// RangeSearch queries the index with the vectors in x.
//	// Returns all vectors with distance < radius.
//	RangeSearch(x []float32, radius float32) (*RangeSearchResult, error)
//
//	// Reset removes all vectors from the index.
//	Reset() error
//
//	// RemoveIDs removes the vectors specified by sel from the index.
//	// Returns the number of elements removed and error.
//	RemoveIDs(sel *IDSelector) (int, error)
//
//	// Delete frees the memory used by the index.
//	Delete()
//}

var defaultConfig = Config{
	Embedding: struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}{
		Provider: "OpenAI",
		Model:    "AdaEmbeddingV2",
	},
	ChunkSpecs: []ChunkSpec{
		{MaxTokens: 512, Overlap: 128},
		{MaxTokens: 1024, Overlap: 256},
	},
}

var ErrNoConfig = errors.New("no config file found")
var ErrAbort = errors.New("abort")

type ChunkSpec struct {
	MaxTokens int `json:"max_tokens"`
	Overlap   int `json:"overlap"`
}

type Config struct {
	Embedding struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	} `json:"embedding"`

	ChunkSpecs []ChunkSpec `json:"chunk_specs"`
}

type Repository struct {
	config Config

	repoPath   string
	ftiPath    string
	configPath string

	embedder llm.Embedder
	chunker  chunkers.Chunker
	index    *OnlineIndex

	ignore *ignore.GitIgnore
}

func NewRepository(repoPath string) (r *Repository, err error) {
	r = &Repository{}

	r.chunker = chunkers.TikToken{}
	r.embedder = &openai.Embedder{
		Client: openai.NewClient(),
		Model:  openai.AdaEmbeddingV2,
	}

	r.repoPath = repoPath
	r.ftiPath = filepath.Join(r.repoPath, ".fti")
	r.configPath = r.ResolveDbPath("config.json")

	if err := r.loadConfig(); err != nil {
		if err != ErrNoConfig {
			return nil, err
		}
	}

	if err := r.loadIgnoreFile(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	r.index, err = NewOnlineIndex(r)

	if err != nil {
		return nil, err
	}

	if err := r.loadIndex(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return r, nil
}

func (r *Repository) RepoPath() string { return r.repoPath }

func (r *Repository) ResolveDbPath(p ...string) string {
	return filepath.Join(r.ftiPath, filepath.Join(p...))
}

func (r *Repository) ResolvePath(p ...string) string {
	return filepath.Join(r.repoPath, filepath.Join(p...))
}

func (r *Repository) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)

	return err == nil
}

func (r *Repository) OpenFile(filePath string) (fs.File, error) {
	return os.OpenFile(filePath, os.O_RDONLY, 0644)
}

func (r *Repository) IsIgnored(name string) bool {
	p := r.RelativeToRoot(name)

	return r.ignore.MatchesPath(p)
}

func (r *Repository) RelativeToRoot(name string) string {
	p, err := filepath.Rel(r.repoPath, name)

	if err != nil {
		return name
	}

	return p
}

func (r *Repository) loadConfig() error {
	if !r.FileExists(r.configPath) {
		return ErrNoConfig
	}

	f, err := r.OpenFile(r.configPath)

	if err != nil {
		return err
	}

	defer f.Close()

	data, err := io.ReadAll(f)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &r.config); err != nil {
		return err
	}

	return nil
}

func (r *Repository) loadIgnoreFile() error {
	ignoreFilePath := r.ResolvePath(".ftiignore")

	if !r.FileExists(ignoreFilePath) {
		return nil
	}

	result, err := ignore.CompileIgnoreFile(ignoreFilePath)

	if err != nil {
		return err
	}

	r.ignore = result

	return nil
}

func (r *Repository) IterateFiles(ctx context.Context) Iterator[FileCursor] {
	files := IterateFiles(ctx, r.repoPath)

	files = Filter(files, func(f FileCursor) bool {
		if f.IsDir() {
			return false
		}

		relPath, err := filepath.Rel(r.ftiPath, f.Path)

		if err == nil && !strings.HasPrefix(relPath, "..") {
			return false
		}

		if r.IsIgnored(f.Path) {
			return false
		}

		return true
	})

	return files
}

func (r *Repository) Init() error {
	// Create the .fti directory
	err := os.Mkdir(r.ftiPath, 0755)
	if err != nil {
		return err
	}

	configData, err := json.MarshalIndent(defaultConfig, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(r.configPath, configData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Update(ctx context.Context) error {
	for it := r.IterateFiles(ctx); it.Next(); {
		f := it.Item()

		if err := r.UpdateFile(ctx, f); err != nil {
			return err
		}
	}

	err := faiss.WriteIndex(r.index.idx, r.ResolveDbPath("index.faiss"))

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateFile(ctx context.Context, f FileCursor) error {
	fmt.Printf("Updating file %s\n", f.Path)

	fh, err := r.OpenFile(f.Path)

	if err != nil {
		return err
	}

	defer fh.Close()

	hasher := sha256.New()
	reader := io.TeeReader(fh, hasher)
	data, err := io.ReadAll(reader)

	if err != nil {
		return err
	}

	h := hasher.Sum(nil)
	fileHash := hex.EncodeToString(h)

	fileDir := r.ResolveDbPath("objects", fileHash)
	metaPath := filepath.Join(fileDir, "meta.json")

	if err := os.MkdirAll(fileDir, 0755); err != nil {
		return err
	}

	meta := &ObjectSnapshotMetadata{
		Path:       f.Path,
		Hash:       fileHash,
		ChunkCount: make([]int, len(r.config.ChunkSpecs)),
	}

	for i, chunkSpec := range r.config.ChunkSpecs {
		img, err := r.updateFileWithSpec(ctx, chunkSpec, fileDir, data)

		if err != nil {
			return err
		}

		meta.ChunkCount[i] = len(img.Chunks)
	}

	serialized, err := json.MarshalIndent(meta, "", "\t")

	if err != nil {
		return err
	}

	if err := os.WriteFile(metaPath, serialized, 0755); err != nil {
		return err
	}

	return nil
}

func (r *Repository) updateFileWithSpec(ctx context.Context, spec ChunkSpec, fileDir string, data []byte) (*ObjectSnapshotImage, error) {
	imagePath := filepath.Join(fileDir, fmt.Sprintf("%dm%d.png", spec.MaxTokens, spec.Overlap))

	chunks, err := r.chunker.SplitTextIntoChunks(ctx, string(data), spec.MaxTokens, spec.Overlap)

	if err != nil {
		return nil, err
	}

	chunksStr := make([]string, len(chunks))

	for i, chunk := range chunks {
		chunksStr[i] = chunk.Content
	}

	embeddings, err := r.embedder.GetEmbeddings(ctx, chunksStr)

	if err != nil {
		return nil, err
	}

	img := &ObjectSnapshotImage{
		Chunks:     chunks,
		Embeddings: embeddings,
	}

	for i, chunk := range chunks {
		emb := embeddings[i]

		chunkPath := filepath.Join(fileDir, fmt.Sprintf("%dm%d.%d.txt", spec.MaxTokens, spec.Overlap, chunk.Index))
		embPath := filepath.Join(fileDir, fmt.Sprintf("%dm%d.%d.f32", spec.MaxTokens, spec.Overlap, chunk.Index))

		if err := os.WriteFile(chunkPath, []byte(chunk.Content), 0644); err != nil {
			return nil, err
		}

		buffer := make([]byte, len(emb.Embeddings)*4)

		for j, f := range emb.Embeddings {
			binary.LittleEndian.PutUint32(buffer[j*4:], math.Float32bits(f))
		}

		if err := os.WriteFile(embPath, buffer, 0644); err != nil {
			return nil, err
		}
	}

	fh, err := os.OpenFile(imagePath, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	if _, err := img.WriteTo(fh); err != nil {
		return nil, err
	}

	if err := r.index.Add(img); err != nil {
		return nil, err
	}

	return img, nil
}

func (r *Repository) Query(ctx context.Context, query string, k int64) ([]OnlineIndexQueryHit, error) {
	embs, err := r.embedder.GetEmbeddings(ctx, []string{query})

	if err != nil {
		return nil, err
	}

	hits, err := r.index.Query(embs[0], k)

	if err != nil {
		return nil, err
	}

	return hits, nil
}

func (r *Repository) loadIndex() error {
	p := r.ResolveDbPath("index.faiss")

	if !r.FileExists(p) {
		return nil
	}

	idx, err := faiss.ReadIndex(p, faiss.IOFlagMmap)

	if err != nil {
		return nil
	}

	if r.index.idx != nil {
		r.index.idx.Delete()
		r.index.idx = nil
	}

	r.index.idx = idx

	return nil
}

type OnlineIndex struct {
	Repository *Repository

	m       sync.RWMutex
	idx     faiss.Index
	mapping map[int64]*OnlineIndexEntry
}

type OnlineIndexEntry struct {
	Index     int64
	Chunk     chunkers.Chunk
	Embedding llm.Embedding
}

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

type OnlineIndexQueryHit struct {
	Entry    *OnlineIndexEntry
	Distance float32
}

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

func (oi *OnlineIndex) putEntry(idx int64, entry *OnlineIndexEntry) error {
	indexPath := oi.Repository.ResolveDbPath("index")

	if err := os.MkdirAll(indexPath, 0644); err != nil {
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

type ObjectSnapshotMetadata struct {
	Path       string `json:"path"`
	Hash       string `json:"hash"`
	ChunkCount []int  `json:"chunk_count"`
}

type ObjectSnapshotImage struct {
	Chunks     []chunkers.Chunk
	Embeddings []llm.Embedding
}

func (osi *ObjectSnapshotImage) ReadFrom(r io.Reader) (int, error) {
	return 0, nil
}

func (osi *ObjectSnapshotImage) WriteTo(w io.Writer) (int, error) {
	img := image.NewRGBA(image.Rect(0, 0, len(osi.Chunks)*25, 25))

	for i, _ := range osi.Chunks {
		embedding := osi.Embeddings[i]
		n := 25

		for x := 0; x < n; x++ {
			for y := 0; y < n; y++ {
				idx := (y*n + x) * 3

				if idx >= len(embedding.Embeddings) {
					continue
				}

				c := color.RGBA{
					R: uint8(embedding.Embeddings[idx+0] * 256.0),
					G: uint8(embedding.Embeddings[idx+1] * 255.0),
					B: uint8(embedding.Embeddings[idx+2] * 255.0),
					A: 255,
				}

				img.Set(i*25+x, y, c)
			}
		}
	}

	err := png.Encode(w, img)
	if err != nil {
		return 0, err
	}

	return 0, nil
}

type Iterator[T any] interface {
	Next() bool
	Item() T
}

type chIterator[T any] struct {
	ch      chan T
	err     <-chan error
	current *T
}

func (it *chIterator[T]) HasNext() bool {
	return it.ch != nil
}

func (it *chIterator[T]) Next() bool {
	if it.ch == nil {
		return false
	}

	select {
	case err, _ := <-it.err:
		it.ch = nil
		it.current = nil

		if err != nil {
			panic(err)
		}

		return false

	case v, ok := <-it.ch:
		if ok {
			it.current = &v
		} else {
			it.ch = nil
			it.current = nil
		}

		return ok
	}
}

func (it *chIterator[T]) Item() T {
	return *it.current
}

func (it *chIterator[T]) Close() error {
	if it.ch != nil {
		close(it.ch)
		it.ch = nil
	}
	return nil
}

type osFileIterator struct {
	chIterator[FileCursor]
}

func (it *osFileIterator) File() FileCursor {
	return it.Item()
}

type filteredIterator[T any] struct {
	src     Iterator[T]
	pred    func(T) bool
	current T
}

func (f *filteredIterator[T]) Next() bool {
	for f.src.Next() {
		if f.pred(f.src.Item()) {
			f.current = f.src.Item()
			return true
		}
	}

	return false
}

func (f *filteredIterator[T]) Item() T {
	return f.current
}

func Filter[IT Iterator[T], T any](it IT, pred func(T) bool) Iterator[T] {
	return &filteredIterator[T]{src: it, pred: pred}
}

func IterateFiles(ctx context.Context, dirPath string) Iterator[FileCursor] {
	ch := make(chan FileCursor)
	errCh := make(chan error)

	go func() {
		defer close(ch)
		defer close(errCh)

		// WalkDir recursively traverses the directory tree rooted at dirPath
		// and sends each file info to the channel ch
		err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			select {
			case <-ctx.Done():
				return ErrAbort
			default:
			}

			// If there's an error, return it
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ErrAbort
			case ch <- FileCursor{
				DirEntry: d,
				Path:     path,
				Err:      err,
			}:
				return nil
			}
		})

		if err != nil && err != ErrAbort {
			errCh <- err
		}
	}()

	return &osFileIterator{
		chIterator: chIterator[FileCursor]{
			ch: ch,
		},
	}
}

type FileCursor struct {
	fs.DirEntry

	Path string
	Err  error
}

type taskContext struct {
	ctx            context.Context
	proc           goprocess.Process
	current, total int
	err            error
	done           chan struct{}
}

func (t *taskContext) Context() context.Context {
	return t.ctx
}

func (t *taskContext) Update(current, total int) {
	t.current = current
	t.total = total
}

func (t *taskContext) Err() error {
	t.Wait()
	return t.err
}

func (t *taskContext) Cancel() {

}

func (t *taskContext) Wait() {
	_, _ = <-t.done
}

func SpawnTask(ctx context.Context, task Task) TaskHandle {
	tc := &taskContext{ctx: ctx}

	parent := goprocessctx.WithContext(ctx)

	tc.proc = goprocess.GoChild(parent, func(proc goprocess.Process) {
		defer proc.CloseAfterChildren()
		defer close(tc.done)

		if err := task.Run(tc); err != nil {
			tc.err = err
		}
	})

	return tc
}

type TaskProgress interface {
	Context() context.Context
	Update(current, total int)
}

type TaskHandle interface {
	Context() context.Context

	Cancel()
	Wait()
	Err() error
}

type TaskFunc func(progress TaskProgress) error

func (f TaskFunc) Run(progress TaskProgress) error {
	return f(progress)
}

type Task interface {
	Run(progress TaskProgress) error
}
