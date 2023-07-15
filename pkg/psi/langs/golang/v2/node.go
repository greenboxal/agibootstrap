package golang

import (
	"go/ast"
	"go/token"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
)

type Node interface {
	project.AstNode

	GoNode() Node
	GoNodeBase() *NodeBase

	ToGoAst() ast.Node
	ToGoExpr() ast.Expr
	ToGoStmt() ast.Stmt
}

type NodeBase struct {
	project.AstNodeBase

	StartTokenPos token.Pos `json:"startTokenPos,omitempty"`
	EndTokenPos   token.Pos `json:"endTokenPos,omitempty"`
}

func (nb *NodeBase) GoNode() Node          { return nb.PsiNode().(Node) }
func (nb *NodeBase) GoNodeBase() *NodeBase { return nb }

func (nb *NodeBase) ToGoExpr() ast.Expr { return nb.GoNode().ToGoAst().(ast.Expr) }
func (nb *NodeBase) ToGoStmt() ast.Stmt { return nb.GoNode().ToGoAst().(ast.Stmt) }
