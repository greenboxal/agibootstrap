package mdlang

import (
	"strings"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/py"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Node interface {
	psi.Node

	Initialize(self Node)

	Ast() ast.Ast
}

type NodeBase[T ast.Ast] struct {
	psi.NodeBase

	node     T
	comments []string
}

func (nb *NodeBase[T]) IsContainer() bool  { return false }
func (nb *NodeBase[T]) IsLeaf() bool       { return false }
func (nb *NodeBase[T]) Ast() ast.Ast       { return nb.node }
func (nb *NodeBase[T]) Comments() []string { return nb.comments }

func (nb *NodeBase[T]) Initialize(self Node) {
	nb.NodeBase.Init(self, "")
}

func (nb *NodeBase[T]) Update() {
	if nb.IsValid() {
		return
	}

	nb.NodeBase.Update()

}

func AstToPsi(parsed ast.Ast) psi.Node {
	n := &NodeBase[ast.Ast]{node: parsed}

	n.Initialize(n)

	if strNode, ok := parsed.(*ast.Str); ok {
		str := py.StringEscape(strNode.S, true)

		if strings.HasPrefix(str, "// TODO:") {
			n.comments = append(n.comments, str)
		}
	}

	return nil
}
