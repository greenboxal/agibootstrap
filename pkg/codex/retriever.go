//go:build selfwip
// +build selfwip

package codex

import (
	"go/ast"

	"github.com/dave/dst"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Retriever struct{}

var dummy dst.Node

func NewRetriever() *Retriever {
	return &Retriever{}
}

// Collects all type references recursively from root node.
func collectTypeRefs(node dst.Node, refs []*dst.Ident) []*dst.Ident {
	switch x := node.(type) {
	case *dst.Ident:
		obj, ok := x.Obj.(*ast.Object)
		if ok && obj != nil && obj.Kind == ast.Typ {
			refs = append(refs, x)
		}
	case *dst.SelectorExpr:
		if ident, ok := x.X.(*dst.Ident); ok && ident.Obj != nil && ident.Obj.Kind == ast.Pkg {
			pkgName := ident.Name
			typeName := x.Sel.Name
			refs = append(refs, &dst.Ident{Name: typeName, Path: pkgName})
		}
	case *dst.ArrayType:
		refs = collectTypeRefs(x.Elt, refs)
	case *dst.ChanType:
		refs = collectTypeRefs(x.Value, refs)
	case *dst.FuncType:
		refs = collectTypeRefs(x.Params, refs)
		if x.Results != nil {
			refs = collectTypeRefs(x.Results, refs)
		}
	case *dst.MapType:
		refs = collectTypeRefs(x.Key, refs)
		refs = collectTypeRefs(x.Value, refs)
	case *dst.StructType:
		if x.Fields != nil {
			for _, f := range x.Fields.List {
				refs = collectTypeRefs(f.Type, refs)
			}
		}
	case *dst.InterfaceType:
		if x.Methods != nil {
			for _, m := range x.Methods.List {
				refs = collectTypeRefs(m.Type, refs)
			}
		}
	case *dst.StarExpr:
		refs = collectTypeRefs(x.X, refs)
	case *dst.Ellipsis:
		if x.Elt != nil {
			refs = collectTypeRefs(x.Elt, refs)
		}
	case *dst.Field:
		if x.Type != nil {
			refs = collectTypeRefs(x.Type, refs)
		}
	case *dst.FieldList:
		for _, f := range x.List {
			refs = collectTypeRefs(f, refs)
		}
	case *dst.ValueSpec:
		if x.Type != nil {
			refs = collectTypeRefs(x.Type, refs)
		}
	case *dst.TypeSpec:
		if x.Type != nil {
			refs = collectTypeRefs(x.Type, refs)
		}
		if x.Comment != nil {
			refs = collectTypeRefs(x.Comment, refs)
		}
	case *dst.File:
		for _, decl := range x.Decls {
			refs = collectTypeRefs(decl, refs)
		}
	case *dst.GenDecl:
		for _, spec := range x.Specs {
			refs = collectTypeRefs(spec, refs)
		}
	}
	return refs
}

func (r *Retriever) Retrieve(root psi.Node) (interface{}, error) {
	// Walk root node Go AST and collect all references
	refs := []*dst.Ident{}

	// Collect all type references
	for _, node := range root.Ast().(*dst.File).Decls {
		refs = collectTypeRefs(node, refs)
	}

	return refs, nil
}
