package fti

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/pkg/errors"
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

import (
	"encoding/json"
)

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

var oai = openai.NewClient()
var embedder = &openai.Embedder{
	Client: oai,
	Model:  openai.AdaEmbeddingV2,
}

type Repository struct {
	repoPath string
	ftiPath  string
	config   Config
}

func NewRepository(repoPath string) *Repository {
	// Make fields private and add getters, take repoPath as argument
	return &Repository{
		repoPath: repoPath,
	}
}

func (r *Repository) RepoPath() string {
	return r.repoPath
}

// Metadata represents the metadata of an object snapshot
type Metadata struct {
	FileName     string `json:"fileName"`
	ObjectHash   string `json:"objectHash"`
	CreationTime string `json:"creationTime"`
	FileSize     int64  `json:"fileSize"`
}

// ObjectSnapshot represents a snapshot of file chunks
type ObjectSnapshot struct {
	Hash       string      `json:"hash"`
	ChunkSize  int         `json:"chunkSize"`
	Overlap    int         `json:"overlap"`
	Embeddings []Embedding `json:"embeddings"`
}

// Embedding represents an embedding vector
type Embedding struct {
	Vector []float32 `json:"vector"`
}

// Chunk represents a chunk of data within a file
type Chunk struct {
	Data string `json:"data"`
}

// Init method initializes a new FTI repository
func (r *Repository) Init() error {
	// Check if the repository is already initialized
	_, err := os.Stat(fmt.Sprintf("%s/.fti", r.RepoPath))
	if !os.IsNotExist(err) {
		return fmt.Errorf("repository is already initialized")
	}

	// Use the config file to set up the repository
	r.config, err = ReadConfigFile(fmt.Sprintf("%s/.fti/config.json", r.RepoPath))
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Create .fti folder
	ftiPath := fmt.Sprintf("%s/.fti", r.RepoPath)
	err = os.Mkdir(ftiPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create .fti folder: %v", err)
	}

	// Create objects folder
	objectsPath := fmt.Sprintf("%s/objects", ftiPath)
	err = os.Mkdir(objectsPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create objects folder: %v", err)
	}

	// Create index folder
	indexPath := fmt.Sprintf("%s/index", ftiPath)
	err = os.Mkdir(indexPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create index folder: %v", err)
	}

	// Write the default config
	defaultConfig := Config{
		ChunkSizes:          []int{1024, 2048, 4096},
		Overlaps:            []int{256, 512, 1024},
		EmbeddingAPI:        "OpenAI",
		EmbeddingDimensions: 768,
	}

	// Convert the default config to JSON
	defaultConfigData, err := json.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %v", err)
	}

	// Write the default config to the config file
	configFilePath := fmt.Sprintf("%s/.fti/config.json", r.RepoPath)
	err = WriteConfigFile(configFilePath, defaultConfigData)
	if err != nil {
		return fmt.Errorf("failed to write default config file: %v", err)
	}

	fmt.Println("Initializing repository at:", r.RepoPath)
	return nil
}

// WriteConfigFile writes the given config data to the specified file
func WriteConfigFile(filepath string, configData []byte) error {
	err := os.WriteFile(filepath, configData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	return nil
}

func (r *Repository) Update() error {
	ftiPath := filepath.Join(r.repoPath, ".fti")
	objectsDir := filepath.Join(ftiPath, "objects")

	// Walk the repository directory to retrieve the list of files
	err := filepath.WalkDir(r.repoPath, func(filePath string, d os.DirEntry, err error) error {
		if !d.IsDir() && d.Name() != "index.bin" {
			// Calculate the file hash
			fileHash, err := calculateFileHash(filePath)
			if err != nil {
				return err
			}

			// Check if the file already exists in the objects directory
			objectPath := filepath.Join(objectsDir, fileHash)
			if _, err := os.Stat(objectPath); !os.IsNotExist(err) {
				return nil
			}

			// Write updated metadata file for current snapshot file
			info, err := os.Stat(filePath)
			if err != nil {
				return err
			}
			metadata := Metadata{
				FileName:     d.Name(),
				ObjectHash:   fileHash,
				CreationTime: info.ModTime().String(),
				FileSize:     info.Size(),
			}

			// Generate chunks and embeddings for each chunk specification
			for _, chunkSize := range r.config.ChunkSizes {
				for _, overlap := range r.config.Overlaps {
					chunks, err := chunkFile(filePath, chunkSize, overlap)
					if err != nil {
						return err
					}

					embeddings, err := generateEmbeddings(chunks)
					if err != nil {
						return err
					}

					// Update object file for current snapshot file
					objPath := filepath.Join(objectPath, fmt.Sprintf("%dm%d.bin", chunkSize, overlap))
					err = UpdateObjectFile(objPath, embeddings)
					if err != nil {
						return err
					}
				}
			}

			err = writeMetadataFile(filepath.Join(objectPath, "metadata.json"), metadata)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	fmt.Println("Updating repository at:", r.RepoPath)
	return nil
}

// Calculate the hash of a file
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Write metadata file for a snapshot
func writeMetadataFile(filepath string, metadata Metadata) error {
	metadataData, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath, metadataData, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// Parse snapshot filename to retrieve snapshot information
func parseSnapshotFilename(filename string) Snapshot {
	// Split the filename by "_" to extract hash, chunk size, and overlap
	split := strings.Split(filename, "_")

	// Retrieve hash from filename
	hash := split[0]

	// Retrieve chunk size from filename
	chunkSizeStr := strings.TrimSuffix(strings.TrimPrefix(split[1], "m"), "b")
	chunkSize, _ := strconv.Atoi(chunkSizeStr)

	// Retrieve overlap from filename
	overlapStr := strings.TrimSuffix(split[2], ".bin")
	overlap, _ := strconv.Atoi(overlapStr)

	return Snapshot{
		Hash:      hash,
		ChunkSize: chunkSize,
		Overlap:   overlap,
	}
}

// UpdateObjectFile updates the object file for a snapshot file
func UpdateObjectFile(objectFilePath string, embeddings []llm.Embedding) error {
	objectFile, err := os.Create(objectFilePath)
	if err != nil {
		return err
	}
	err = binary.Write(objectFile, binary.LittleEndian, embeddings)
	if err != nil {
		return err
	}
	objectFile.Close()

	return nil
}

// Snapshot represents the structure of a snapshot file
type Snapshot struct {
	Hash      string `json:"hash"`
	ChunkSize int    `json:"chunkSize"`
	Overlap   int    `json:"overlap"`
}

func generateEmbeddings(chunks []string) ([]llm.Embedding, error) {

	// Add your code here to generate embeddings for the given chunks
	fmt.Println("Generating embeddings for chunks:", chunks)
	return nil, nil
}

func chunkFile(filepath string, chunkSize int, overlap int) ([]string, error) {
	fileData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	fileSize := len(fileData)

	chunks := make([]string, 0)

	for i := 0; i < fileSize-chunkSize; i += chunkSize - overlap {
		end := i + chunkSize
		if end > fileSize {
			end = fileSize
		}

		chunk := fileData[i:end]
		chunks = append(chunks, string(chunk))
	}

	return chunks, nil
}

// Query method enables users to query the FTI repository
func (r *Repository) Query(query string) error {
	// Implement query logic based on the design
	objectsPath := filepath.Join(r.ftiPath, "objects")
	indexPath := filepath.Join(r.ftiPath, "index", "inverseIndex.bin")
	queryEmbedding, err := generateEmbeddings([]string{query})
	if err != nil {
		return err
	}
	// Perform inverse index lookup
	objectHash, chunkIndex, _, err := r.lookupInverseIndex(queryEmbedding[0], indexPath)
	if err != nil {
		return err
	}
	// Retrieve object snapshot
	objectSnapshot, err := r.retrieveObjectSnapshot(objectHash, objectsPath)
	if err != nil {
		return err
	}
	// Retrieve metadata for the object snapshot
	metadata, err := r.loadMetadata(objectHash, objectsPath)
	if err != nil {
		return err
	}
	// Extract the relevant chunk from the object snapshot
	chunk := objectSnapshot.Embeddings[chunkIndex]
	// Return the retrieved information
	fmt.Println("Query result:")
	fmt.Println("Object Hash:", objectSnapshot.Hash)
	fmt.Println("Metadata:", metadata)
	fmt.Println("Chunk:", chunk)
	return nil
}

func (r *Repository) lookupInverseIndex(embedding llm.Embedding, path string) (string, int64, int64, error) {
	// Read the inverse index file
	indexData, err := ioutil.ReadFile(path)
	if err != nil {
		return "", 0, 0, err
	}

	// Convert the index entries into a slice of Entry structs
	entries := make([]Entry, 0)
	err = binary.Read(bytes.NewReader(indexData), binary.LittleEndian, &entries)
	if err != nil {
		return "", 0, 0, err
	}

	// Perform a binary search to find the corresponding entry for the embedding
	low := 0
	high := len(entries) - 1
	var embeddingBytes []byte
	var ok bool
	if embeddingBytes, ok = embedding.(bytes.Buffer); !ok {
		return "", 0, 0, errors.New("embedding is not of type bytes.Buffer")
	}

	for low <= high {
		mid := (low + high) / 2
		if bytes.Compare(embeddingBytes, entries[mid].Identifier) == 0 {
			return entries[mid].ObjectHash, entries[mid].ChunkIndex, entries[mid].ChunkOffset, nil
		} else if bytes.Compare(embeddingBytes, entries[mid].Identifier) < 0 {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	// Return an error if the embedding is not found in the inverse index
	return "", 0, 0, errors.New("embedding not found in the inverse index")
}

// Entry represents an entry in the inverse index
type Entry struct {
	Identifier  []byte
	ObjectHash  string
	ChunkIndex  int64
	ChunkOffset int64
}

// Implement the retrieveObjectSnapshot method based on the design
func (r *Repository) retrieveObjectSnapshot(hash interface{}, path string) (ObjectSnapshot, error) {
	objectPath := filepath.Join(path, hash.(string))
	snapshots, err := ioutil.ReadDir(objectPath)
	if err != nil {
		return ObjectSnapshot{}, err
	}

	snapshotInfo := parseSnapshotFilename(snapshots[0].Name())

	// Iterate over the snapshots to retrieve the embedding data
	var embeddings []Embedding
	for _, snapshot := range snapshots {
		if snapshot.IsDir() {
			continue
		}

		snapshotFilename := snapshot.Name()

		// Read the embedding data from the snapshot file
		snapshotFilePath := filepath.Join(objectPath, snapshotFilename)
		snapshotFile, err := os.Open(snapshotFilePath)
		if err != nil {
			return ObjectSnapshot{}, err
		}
		defer snapshotFile.Close()

		// Read the embeddings from the snapshot file
		var snapshotEmbeddings []Embedding
		err = binary.Read(snapshotFile, binary.LittleEndian, &snapshotEmbeddings)
		if err != nil {
			return ObjectSnapshot{}, err
		}
		embeddings = append(embeddings, snapshotEmbeddings...)
	}

	return ObjectSnapshot{
		Hash:       hash.(string),
		ChunkSize:  snapshotInfo.ChunkSize,
		Overlap:    snapshotInfo.Overlap,
		Embeddings: embeddings,
	}, nil
}

// Implement the loadMetadata method based on the design
func (r *Repository) loadMetadata(hash interface{}, path string) (Metadata, error) {
	metadataFilePath := filepath.Join(path, hash.(string), "metadata.json")
	metadataData, err := ioutil.ReadFile(metadataFilePath)
	if err != nil {
		return Metadata{}, err
	}

	var metadata Metadata
	err = json.Unmarshal(metadataData, &metadata)
	if err != nil {
		return Metadata{}, err
	}

	return metadata, nil
}

// Config represents the configuration parameters for the FTI repository
type Config struct {
	ChunkSizes          []int  `json:"ChunkSizes"`
	Overlaps            []int  `json:"Overlaps"`
	EmbeddingAPI        string `json:"EmbeddingAPI"`
	EmbeddingDimensions int    `json:"EmbeddingDimensions"`
}

// ReadConfigFile reads the config file and performs necessary setup based on the configuration
func ReadConfigFile(filepath string) (Config, error) {
	configData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
