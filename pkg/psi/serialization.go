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

	Attributes map[string]interface{} `json:"attr,omitempty"`

	Data cidlink.Link `json:"data,omitempty"`
}

type NodeSnapshot interface {
	Node() Node

	CommittedVersion() int64
	CommittedLink() ipld.Link

	LastFenceID() uint64
}

func UpdateNodeSnapshot(node Node, fn NodeSnapshot) { node.PsiNodeBase().setLastSnapshot(fn) }
func GetNodeSnapshot(node Node) NodeSnapshot        { return node.PsiNodeBase().getLastSnapshot() }
