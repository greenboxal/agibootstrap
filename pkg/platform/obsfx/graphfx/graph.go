package graphfx

import (
	"gonum.org/v1/gonum/graph"

	"github.com/greenboxal/agibootstrap/pkg/platform/obsfx"
)

type GraphChangeEvent[K comparable, N Node, E Edge] interface {
	Graph() ObservableGraph[K, N, E]
}

type NodeChangeEvent[K comparable, N Node, E Edge] struct {
	G            ObservableGraph[K, N, E]
	Key          K
	ValueAdded   N
	ValueRemoved N
	WasAdded     bool
	WasRemoved   bool
}

func (n *NodeChangeEvent[K, N, E]) Graph() ObservableGraph[K, N, E] {
	return n.G
}

type EdgeChangeEvent[K comparable, N Node, E Edge] struct {
	G            ObservableGraph[K, N, E]
	ValueAdded   E
	ValueRemoved E
	WasAdded     bool
	WasRemoved   bool
}

func (n *EdgeChangeEvent[K, N, E]) Graph() ObservableGraph[K, N, E] {
	return n.G
}

type GraphListener[K comparable, N Node, E Edge] interface {
	OnGraphChanged(ev GraphChangeEvent[K, N, E])
}

type Node = graph.Node
type Edge = graph.Edge
type WeightedEdge = graph.WeightedEdge

type ObservableGraph[K comparable, N Node, E Edge] interface {
	graph.Graph
	obsfx.Observable

	AddGraphListener(listener GraphListener[K, N, E])
	RemoveGraphListener(listener GraphListener[K, N, E])
}

type WeightedGraph[K comparable, N Node, E WeightedEdge] interface {
	ObservableGraph[K, N, E]

	graph.Weighted
}

type DirectedGraph[K comparable, N Node, E Edge] interface {
	ObservableGraph[K, N, E]

	graph.Directed
}

type UndirectedGraph[K comparable, N Node, E Edge] interface {
	ObservableGraph[K, N, E]

	graph.WeightedBuilder
}

type WeightedDirectedGraph[K comparable, N Node, E WeightedEdge] interface {
	DirectedGraph[K, N, E]
	WeightedGraph[K, N, E]

	graph.WeightedDirected
}

type WeightedUndirectedGraph[K comparable, N Node, E WeightedEdge] interface {
	UndirectedGraph[K, N, E]
	WeightedGraph[K, N, E]

	graph.WeightedUndirected
}

type MutableGraph[K comparable, N Node, E Edge] interface {
	ObservableGraph[K, N, E]

	graph.Builder
}

type MutableWeightedGraph[K comparable, N Node, E WeightedEdge] interface {
	MutableGraph[K, N, E]
	WeightedGraph[K, N, E]

	graph.WeightedBuilder
}

type MutableDirectedGraph[K comparable, N Node, E Edge] interface {
	MutableGraph[K, N, E]
	DirectedGraph[K, N, E]

	graph.DirectedBuilder
}

type MutableUndirectedGraph[K comparable, N Node, E Edge] interface {
	MutableGraph[K, N, E]
	UndirectedGraph[K, N, E]

	graph.UndirectedBuilder
}

type MutableWeightedDirectedGraph[K comparable, N Node, E WeightedEdge] interface {
	MutableDirectedGraph[K, N, E]
	WeightedDirectedGraph[K, N, E]

	graph.DirectedWeightedBuilder
}

type MutableWeightedUndirectedGraph[K comparable, N Node, E WeightedEdge] interface {
	MutableUndirectedGraph[K, N, E]
	WeightedDirectedGraph[K, N, E]

	graph.UndirectedWeightedBuilder
}
