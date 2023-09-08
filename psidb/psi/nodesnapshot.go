package psi

import (
	"context"

	"github.com/ipld/go-ipld-prime"
)

type NodeSnapshot interface {
	ID() int64
	Node() Node
	Path() Path

	CommitVersion() int64
	CommitLink() ipld.Link

	LastFenceID() uint64
	FrozenNode() *FrozenNode
	FrozenEdges() []*FrozenEdge

	OnBeforeInitialize(node Node)
	OnAfterInitialize(node Node)
	OnInvalidated()
	OnUpdated(ctx context.Context) error

	Resolve(ctx context.Context, path Path) (Node, error)

	OnEdgeAdded(added Edge)
	OnEdgeRemoved(removed Edge)

	OnAttributeChanged(key string, added any)
	OnAttributeRemoved(key string, removed any)

	OnParentChange(newParent Node)

	Lookup(key EdgeReference) Edge
}

func GetNodeSnapshot(node Node) NodeSnapshot { return node.PsiNodeBase().GetSnapshot() }
