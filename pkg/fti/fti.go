package fti

/*
# File Tree Indexing Design Document

## Introduction

A File Tree Index (FTI) serves as an efficient and effective vector index for a given directory, herein referred to as a repository. The FTI system allows us to encapsulate the essence of files within the repository using embedding vectors. Each file in the repository is chunked according to a specific configuration, comprising a chunk size and overlap. The system uses external APIs such as OpenAI's Embeddings API to generate an embedding vector for each chunk.

## Physical Layout

The FTI system uses a standard structure stored within a `.fti` folder located at the root of the FTI repository. The structure of the `.fti` folder includes two main components:

```
/path/to/my/repo/.fti/
  - objects/
  - <object hash>/
  - <chunkSize>m<overlap>.bin
  - index/

```

### Objects

Objects are snapshots of given content stored in a content-addressable fashion. Each object is preserved under `.fti/objects` as a directory, named after the hash of the data. For every chunking specification, there exists an object snapshot file termed `<chunkSize>m<overlap>.bin`.

These snapshot objects are visualized as 2D square images with the embeddings of each chunk distributed as RGB squares inside each chunk.

### Inverse Index

The Inverse Index resides under the `.fti/index/` directory and represents a crucial component of the FTI system. It stores a linear binary tree that maps int64 values to a pair of object hash and chunk index. This binary file is optimized for in-memory retrieval, providing swift and accurate data lookup.
*/

import (
	"fmt"
	"os"
)

// Repository type
type Repository struct {
	RepoPath string
	FTIPath  string
}

// Init method initializes a new FTI repository
func (r *Repository) Init() error {
	ftiPath := fmt.Sprintf("%s/.fti", r.RepoPath)
	err := os.Mkdir(ftiPath, os.ModePerm) // Creates the .fti folder
	if err != nil {
		return err
	}
	// Objects folder
	err = os.Mkdir(fmt.Sprintf("%s/objects", ftiPath), os.ModePerm)
	if err != nil {
		return err
	}
	// Index folder
	err = os.Mkdir(fmt.Sprintf("%s/index", ftiPath), os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Println("Initializing repository at:", r.RepoPath)
	return nil
}

// Update method updates the FTI repository
func (r *Repository) Update() error {
	ftiPath := fmt.Sprintf("%s/.fti", r.RepoPath)
	indexDir := fmt.Sprintf("%s/index", ftiPath)

	// Retrieve the list of chunkSize and overlap binary snapshot files from the `index` directory
	files, err := ioutil.ReadDir(indexDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		// Read snapshot file into memory
		snapshotFilePath := fmt.Sprintf("%s/%s", indexDir, f.Name())
		snapshotFile, err := os.Open(snapshotFilePath)
		if err != nil {
			return err
		}
		snapshotData, err := ioutil.ReadAll(snapshotFile)
		if err != nil {
			return err
		}

		// Unmarshal the snapshot data into a map to get hash and chunk specification
		var snapshot map[string]interface{}
		err = json.Unmarshal(snapshotData, &snapshot)
		if err != nil {
			return err
		}
		hash := snapshot["hash"].(string)
		chunkSize := int(snapshot["chunkSize"].(float64))
		overlap := int(snapshot["overlap"].(float64))
		snapshotFile.Close()

		// Generate embeddings for current snapshot file
		embeddings, err := GenerateEmbeddings(fmt.Sprintf("%s/%d_%d.bin", ftiPath, chunkSize, overlap))
		if err != nil {
			return err
		}

		// Update object file for current snapshot file
		objectFilePath := fmt.Sprintf("%s/objects/%s/%d_%d.bin", ftiPath, hash, chunkSize, overlap)
		objectFile, err := os.Create(objectFilePath)
		if err != nil {
			return err
		}
		err = binary.Write(objectFile, binary.LittleEndian, embeddings)
		if err != nil {
			return err
		}
		objectFile.Close()
	}

	fmt.Println("Updating repository at:", r.RepoPath)
	return nil
}

// Query method enables users to query the FTI repository
func (r *Repository) Query(query string) error {
	// query logic goes here
	fmt.Println("Querying repository at:", r.RepoPath, "with query:", query)
	return nil
}
