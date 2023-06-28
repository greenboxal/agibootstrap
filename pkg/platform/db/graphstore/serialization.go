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
