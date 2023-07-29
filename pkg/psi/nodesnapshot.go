package psi

import "github.com/ipld/go-ipld-prime"

type NodeSnapshot interface {
	ID() int64
	Node() Node
	Path() Path

	CommitVersion() int64
	CommitLink() ipld.Link

	LastFenceID() uint64
	FrozenNode() *FrozenNode
	FrozenEdges() []*FrozenEdge
}

func GetNodeSnapshot(node Node) NodeSnapshot { return node.PsiNodeBase().GetSnapshot() }
