package ford

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/stdlib"
)

const EdgeKindDependsOn psi.EdgeKind = "depends_on"
const EdgeKindRequires psi.EdgeKind = "requires"
const EdgeKindProvides psi.EdgeKind = "provides"

type Network struct {
	stdlib.NodeCollection[*Node]
}

func NewNetwork() *Network {
	n := &Network{}

	n.Init(n, "")

	return n
}

func (n *Network) DependencyGraph() graph.Directed {
	g := simple.NewDirectedGraph()

	for it := n.Iterator(); it.Next(); {
		n := it.Value()

		gn := g.Node(n.ID())

		if gn == nil {
			g.AddNode(n)
		}

		for it := n.DependsOn.Iterator(); it.Next(); {
			dep := it.Value()

			g.SetEdge(g.NewEdge(n, dep))
		}
	}

	return g
}

type EdgeCollection[T psi.Node] struct {
	stdlib.NodeCollection[T]
	name string
}

func (c *EdgeCollection[T]) PsiNodeName() string { return c.name }

func (c *EdgeCollection[T]) Init(name psi.EdgeKind) {
	c.name = name.String()

	c.NodeCollection.Init(c, "")
}

type Node struct {
	psi.NodeBase

	Name     string
	Requires []string
	Provides []string

	DependsOn EdgeCollection[*Node]
}

func (n *Node) PsiNodeName() string { return n.Name }

func (n *Node) Init() {
	n.DependsOn.Init(EdgeKindDependsOn)
	n.DependsOn.SetParent(n)
}
