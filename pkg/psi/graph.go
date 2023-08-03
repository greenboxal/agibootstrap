package psi

import (
	"context"
)

type Graph interface {
	Add(n Node)
	Remove(n Node)

	NextEdgeID() EdgeID

	ResolveNode(ctx context.Context, path Path) (n Node, err error)
	ListNodeEdges(ctx context.Context, path Path) (result []*FrozenEdge, err error)
}
