package graphstore

import (
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type IndexedGraph struct {
	ID    uuid.UUID
	Nodes []FrozenNode
}

type IndexedNode struct {
	Path    string
	Version uint64
	Current cid.Cid
}

type IndexedEdge struct {
	Version uint64
}

type FrozenNode struct {
	Cid  cid.Cid
	Addr cid.Cid

	ID       psi.NodeID
	ParentID psi.NodeID
	Type     psi.NodeType

	Attr map[string]any
}

type FrozenEdge struct {
	Cid  cid.Cid
	Addr cid.Cid

	ID   psi.EdgeID
	Type psi.EdgeKind

	From psi.NodeID
	To   psi.NodeID

	Attr map[string]any
}

func init() {
	// TODO: Write design document skeleton for "graphstore serialization". This will be used to serialize a graph so it can be stored in the disk using a KV embedded database.
}

// Design document skeleton for "graphstore serialization":
/*
Serialization is the process of converting the in-memory representation of a graph into a format that can be stored in a disk-based Key-Value (KV) embedded database.
This design document defines the key components and considerations for serializing a graph in the graphstore package.

1. Graph Structure:
   - The graph structure consists of nodes and edges.
   - Each node has an ID, UUID, type, parent ID, and attributes.
   - Each edge has an ID, type, origin node ID, destination node ID, and attributes.

2. Data Formats:
   - The graph structure can be serialized using efficient data formats like JSON, Protocol Buffers, or binary formats.
   - The data format chosen should support efficient read and write operations and minimize storage space.

3. Encoding and Decoding:
   - The graph structure needs to be encoded into the chosen data format for storage and decoded for in-memory retrieval.
   - Encoding involves converting the graph structure into a byte array using the chosen data format.
   - Decoding involves parsing the byte array into the graph structure using the chosen data format.

4. Key-Value Storage:
   - The serialized graph can be stored in a disk-based KV embedded database.
   - The key should uniquely identify each serialized graph or its components (nodes, edges).
   - The value should contain the serialized representation of the graph or its components.

5. Versioning:
   - Consider adding versioning support to track updates and changes to the graph.
   - Versioning can be achieved by associating a version number with each serialized graph or its components.
   - Updates to the graph can be tracked by incrementing the version number.

6. Error Handling and Recovery:
   - Define error handling mechanisms for serialization and deserialization failures.
   - Implement recovery mechanisms to handle corrupted or incomplete serialized data.

7. Performance Optimization:
   - Consider optimizing the serialization and deserialization processes for performance.
   - Use efficient data structures and algorithms to minimize the time and space complexity.

8. Integration with KV Embedded Database:
   - Integrate the serialization and deserialization processes with the chosen KV embedded database's APIs.
   - Implement CRUD operations (Create, Read, Update, Delete) for the serialized graph or its components.

9. Unit Testing:
   - Develop test cases to ensure the correctness and reliability of the serialization and deserialization processes.
   - Test different scenarios, including edge cases and error conditions.

10. Documentation:
    - Provide extensive documentation explaining the serialization and deserialization processes, including usage guidelines and examples.

Note: The design document should be continuously updated and refined as the implementation progresses to accommodate any changes or new requirements.
*/
