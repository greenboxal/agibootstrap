package mdlang

import (
	"github.com/gomarkdown/markdown/ast"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Node interface {
	psi.Node

	Initialize(self Node)

	Ast() ast.Node
}

type NodeBase[T ast.Node] struct {
	psi.NodeBase

	node T
}

func (nb *NodeBase[T]) IsContainer() bool { return nb.node.AsContainer() != nil }
func (nb *NodeBase[T]) IsLeaf() bool      { return nb.node.AsLeaf() != nil }
func (nb *NodeBase[T]) Ast() ast.Node     { return nb.node }

func (nb *NodeBase[T]) Comments() []string { return nil }

func (nb *NodeBase[T]) Initialize(self Node) {
	nb.NodeBase.Init(self, "")
}

func AstToPsi(node ast.Node) (result psi.Node) {
	containerStack := make([]psi.Node, 0, 16)

	ast.WalkFunc(node, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			wrapped := &NodeBase[ast.Node]{
				node: node,
			}

			wrapped.Init(wrapped, "")

			if len(containerStack) > 0 {
				wrapped.SetParent(containerStack[len(containerStack)-1])
			}

			if node.AsContainer() != nil {
				containerStack = append(containerStack, wrapped)
			}
		} else if node.AsContainer() != nil {
			current := containerStack[len(containerStack)-1]

			containerStack = containerStack[:len(containerStack)-1]

			if len(containerStack) == 0 {
				result = current
			}
		}

		return ast.GoToNext
	})

	return
}
