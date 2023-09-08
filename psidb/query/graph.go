package query

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/iterator"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type nodeWrapper struct {
	Node psi.Node
}

func (n nodeWrapper) ID() int64 { return n.Node.ID() }

type edgeWrapper struct {
	g    *GraphWrapper
	Edge psi.Edge
}

func (e edgeWrapper) From() graph.Node         { return e.g.Add(e.Edge.From()) }
func (e edgeWrapper) To() graph.Node           { return e.g.Add(e.Edge.To()) }
func (e edgeWrapper) ReversedEdge() graph.Edge { return nil }

type GraphWrapper struct {
	Graph psi.Graph

	idMap map[int64]graph.Node
}

func (g *GraphWrapper) Node(id int64) graph.Node {
	return g.idMap[id]
}

func (g *GraphWrapper) Nodes() graph.Nodes {
	return iterator.NewNodes(g.idMap)
}

func (g *GraphWrapper) HasEdgeBetween(xid, yid int64) bool {
	u := g.idMap[xid].(*nodeWrapper).Node

	for edges := u.Edges(); edges.Next(); {
		to := edges.Value().To()

		if to.ID() == yid {
			return true
		}
	}

	return false
}

func NewGraphWrapper(g psi.Graph) *GraphWrapper {
	return &GraphWrapper{
		Graph: g,

		idMap: make(map[int64]graph.Node),
	}
}

func (g *GraphWrapper) Add(n psi.Node) graph.Node {
	g.idMap[n.ID()] = n

	return nodeWrapper{Node: n}
}

func (g *GraphWrapper) From(id int64) graph.Nodes {
	u := g.idMap[id].(*nodeWrapper).Node

	edges := iterators.ToSlice(iterators.Map(u.Edges(), func(e psi.Edge) graph.Node {
		return g.Add(e.To())
	}))

	return iterator.NewOrderedNodes(edges)
}

func (g *GraphWrapper) Edge(uid, vid int64) graph.Edge {
	u := g.idMap[uid].(*nodeWrapper).Node

	v := iterators.FirstOrNull(iterators.Filter(u.Edges(), func(e psi.Edge) bool {
		to := e.To()

		return to.ID() == vid
	}))

	if v == nil {
		return nil
	}

	return edgeWrapper{
		g:    g,
		Edge: v,
	}
}
