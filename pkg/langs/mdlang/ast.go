package mdlang

import (
	"github.com/gomarkdown/markdown/ast"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Node interface {
	psi.Node

	Initialize(self Node)
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

type CodeBlock struct{ NodeBase[*ast.CodeBlock] }

func AstToPsi(node ast.Node) (psi.Node, error) {
	var result Node

	switch node := node.(type) {
	case *ast.CodeBlock:
		result = &CodeBlock{}
	default:
		return nil, errors.Errorf("unsupported node type: %T", node)
	}

	result.Initialize(result)

	return result, nil
}
