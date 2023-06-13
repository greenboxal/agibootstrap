//go:build prod
// +build prod

package codex

import (
	"go/ast"

	"github.com/dave/dst"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Retriever struct {
}

var dummy dst.Node

func NewRetriever() *Retriever {
	return &Retriever{}
}

func (r *Retriever) Retrieve(root psi.Node) (interface{}, error) {
	// Walk root node Go AST and collect all references
	refs := []*dst.Ident{}
	decorator.Walk(root.Ast().(*ast.File), func(node dst.Node) bool {
		switch x := node.(type) {
		case *dst.Ident:
			// Checking for references
			if x.Obj != nil {
				refs = append(refs, x)
			}
		}
		return true
	})

	return refs, nil
}
