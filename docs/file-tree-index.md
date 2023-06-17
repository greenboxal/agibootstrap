# File Tree Indexing Design Document
## Introduction
A File Tree Index (FTI) serves as an efficient and effective vector index for a given directory, herein referred to as a repository. The FTI system allows us to encapsulate the essence of files within the repository using embedding vectors. Each file in the repository is chunked according to a specific configuration, comprising a chunk size and overlap. The system uses external APIs such as OpenAI's Embeddings API to generate an embedding vector for each chunk.
## Physical Layout
The FTI system uses a standard structure stored within a `.fti` folder located at the root of the FTI repository. The structure of the `.fti` folder includes two main components:
### Objects
Objects are snapshots of given content stored in a content-addressable fashion. Each object is preserved under `.fti/objects` as a directory, named after the hash of the data. For every chunking specification, there exists an object snapshot file termed `<chunkSize>m<overlap>.bin`. These snapshot objects are visualized as 2D square images with the embeddings of each chunk distributed as RGB squares inside each chunk.
### Inverse Index
The Inverse Index resides under the `.fti/index/` directory and represents a crucial component of the FTI system. It stores a linear binary tree that maps int64 values to a pair of object hash and chunk index. This binary file is optimized for in-memory retrieval, providing swift and accurate data lookup.
## Command-Line Interface (CLI): FTI
The FTI system includes a CLI tool named `fti` to provide user-friendly interactions. This tool supports several commands: `init`, `update`, and `query`.
### `fti init`
The `init` command initializes a new FTI repository. It sets up the `.fti` directory at the root of the repository and prepares the structure to accommodate future index operations.
### `fti update`
The `update` command is responsible for updating the FTI repository. It scans the repository for any new or modified files, breaks them into chunks, generates the embedding vectors, and updates the object and Inverse Index files.
### `fti query`
The `query` command enables users to query the FTI repository. Upon entering a query, the system retrieves the relevant file information based on the embedding vectors in the FTI. The Inverse Index ensures efficient data retrieval.
## Implementation Plan
### Step 1: Implement CLI Structure
We will start by creating the structure of our CLI tool using the Cobra library. This structure will consist of the root command and the `init`, `update`, and `query` subcommands.
### Step 2: Implement Utils
Next, we will write utility functions that we will use across the application. This will include file handling operations, object management, and index handling operations.
### Step 3: Implement Command Logic
After our utility functions are ready, we will implement the logic for the `init`, `update`, and `query` commands in their respective files.
### Step 4: Test and Debug
Finally, we will write unit tests for our utility functions and commands to ensure they work as expected. We will also manually test our CLI tool and fix any bugs that we encounter. Upon completion, we will have a fully functional File Tree Indexing system ready for use. The FTI will streamline the process of managing large directories and retrieving data efficiently.