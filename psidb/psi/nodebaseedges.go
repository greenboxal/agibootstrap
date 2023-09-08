package psi

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
)

func (n *NodeBase) Edges() EdgeIterator {
	return iterators.Map(n.edges.Iterator(), func(t collectionsfx.KeyValuePair[EdgeKey, Edge]) Edge {
		return t.Value
	})
}

func (n *NodeBase) UpsertEdge(edge Edge) {
	n.edges.Set(edge.Key().GetKey(), edge)
}

func (n *NodeBase) SetEdge(key EdgeReference, to Node) {
	e := n.GetEdge(key.GetKey())

	if e == nil {
		e = NewSimpleEdge(key, n, to)
	} else {
		if e.To() == to {
			return
		}

		e = e.ReplaceTo(to)
	}

	n.UpsertEdge(e)
}

func (n *NodeBase) UnsetEdge(key EdgeReference) {
	k := key.GetKey()

	_, ok := n.edges.Get(k)

	if !ok {
		return
	}

	n.edges.Remove(k)
}

func (n *NodeBase) GetEdge(key EdgeReference) Edge {
	v, ok := n.edges.Get(key.GetKey())

	if !ok || v == nil {
		if n.snap != nil {
			return n.snap.Lookup(key)
		}
	}

	return v
}
