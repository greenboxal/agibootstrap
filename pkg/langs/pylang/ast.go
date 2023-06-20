package pylang

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

func NewNodeFor(node ast.Ast) psi.Node {
	n := &NodeBase[ast.Ast]{node: node}

	n.Initialize(n)

	return n
}

func AstToPsi(parsed ast.Ast) (result psi.Node) {
	result = NewNodeFor(parsed)

	ast.Walk(parsed, func(node ast.Ast) bool {
		if node == parsed {
			return true
		} else {
			wrapped := AstToPsi(node)
			wrapped.SetParent(result)

			return false
		}
	})

	if strNode, ok := parsed.(*ast.Str); ok {
		str := py.StringEscape(strNode.S, true)

		if strings.HasPrefix(str, "// TODO:") {
			result.(*NodeBase[ast.Ast]).comments = append(result.(*NodeBase[ast.Ast]).comments, str)
		}
	}

	return
}
