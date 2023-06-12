package astparser

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

// Parse parses Go code into an AST.
func Parse(code string) (*ast.File, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// ToCode converts an AST back to Go code.
func ToCode(node ast.Node) (string, error) {
	fset := token.NewFileSet()
	var buf bytes.Buffer
	err := format.Node(&buf, fset, node)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
