package psi

import (
	"time"

	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

type FrozenGraph struct {
	Nodes []*FrozenNode
	Edges []*FrozenEdge
}

type FrozenNode struct {
	Cid cidlink.Link `json:"cid,omitempty"`

	Index   int64 `json:"index"`
	Version int64 `json:"version"`

	UUID NodeID   `json:"uuid"`
	Type NodeType `json:"type"`

	Attributes map[string]interface{} `json:"attr,omitempty"`
}

type FrozenEdge struct {
	Cid cidlink.Link `json:"cid,omitempty"`

	Key EdgeKey `json:"key"`

	From NodeID `json:"from"`
	To   NodeID `json:"to"`

	Attributes map[string]interface{} `json:"attr,omitempty"`
}

type NodeSnapshot struct {
	Link      ipld.Link
	Version   int64
	Timestamp time.Time
}

func UpdateNodeSnapshot(node Node, fn *NodeSnapshot) { node.setLastSnapshot(fn) }
func GetNodeSnapshot(node Node) *NodeSnapshot        { return node.getLastSnapshot() }
