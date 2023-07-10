package psi

import (
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

type FrozenGraph struct {
	Nodes []*FrozenNode
	Edges []*FrozenEdge
}

type FrozenNode struct {
	Index   int64 `json:"index"`
	Version int64 `json:"version"`

	Path Path   `json:"path"`
	Type string `json:"type"`

	Edges      []cidlink.Link         `json:"edges,omitempty"`
	Attributes map[string]interface{} `json:"attr,omitempty"`

	Data *cidlink.Link `json:"link,omitempty"`
}

type FrozenEdge struct {
	Key EdgeKey `json:"key"`

	FromPath Path          `json:"from_path"`
	ToPath   *Path         `json:"to_path"`
	ToLink   *cidlink.Link `json:"to_link"`
	ToIndex  int64         `json:"to_index"`

	Attributes map[string]interface{} `json:"attr,omitempty"`

	Data cidlink.Link `json:"data,omitempty"`
}

type NodeSnapshot interface {
	ID() int64
	Node() Node
	Path() Path

	CommitVersion() int64
	CommitLink() ipld.Link

	LastFenceID() uint64
	FrozenNode() *FrozenNode
}

func UpdateNodeSnapshot(node Node, fn NodeSnapshot) { node.PsiNodeBase().SetSnapshot(fn) }
func GetNodeSnapshot(node Node) NodeSnapshot        { return node.PsiNodeBase().GetSnapshot() }
