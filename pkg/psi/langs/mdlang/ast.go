package mdlang

import (
	"context"
	"fmt"

	"github.com/gomarkdown/markdown/ast"
	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Node interface {
	psi.Node

	Initialize(self Node)

	Ast() ast.Node
}

type NodeBase[T ast.Node] struct {
	psi.NodeBase

	node     T
	comments []string
}

func (nb *NodeBase[T]) IsContainer() bool  { return nb.node.AsContainer() != nil }
func (nb *NodeBase[T]) IsLeaf() bool       { return nb.node.AsLeaf() != nil }
func (nb *NodeBase[T]) Ast() ast.Node      { return nb.node }
func (nb *NodeBase[T]) Comments() []string { return nb.comments }

func (nb *NodeBase[T]) String() string {
	return fmt.Sprintf("%T(%d)", nb.node, nb.ID())
}

func (nb *NodeBase[T]) Initialize(self Node) {
	nb.NodeBase.Init(self)
}

func (nb *NodeBase[T]) OnUpdate(ctx context.Context) error {
	if nb.IsValid() {
		return nil
	}

	if err := nb.NodeBase.OnUpdate(ctx); err != nil {
		return nil
	}

	if c := nb.node.AsContainer(); c != nil {
		c.SetChildren(lo.Map(nb.Children(), func(n psi.Node, _ int) ast.Node {
			mdn := n.(Node).Ast()
			mdn.SetParent(nb.node)
			return mdn
		}))
	}
	return nil
}

func AstToPsi(node ast.Node) (result psi.Node) {
	containerStack := make([]psi.Node, 0, 16)

	ast.WalkFunc(node, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			wrapped := &NodeBase[ast.Node]{
				node: node,
			}

			wrapped.Init(wrapped)

			if len(containerStack) > 0 {
				wrapped.SetParent(containerStack[len(containerStack)-1])
			}

			if c := node.AsContainer(); c != nil {
				containerStack = append(containerStack, wrapped)
				wrapped.comments = append(wrapped.comments, string(c.Literal))
			}

			if f := node.AsLeaf(); f != nil {
				wrapped.comments = append(wrapped.comments, string(f.Literal))
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
