package psi

import (
	"bytes"
	"errors"
	"go/parser"
	"go/token"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

type SourceFile struct {
	name string
	dec  *decorator.Decorator
	fset *token.FileSet

	root   Node
	parsed *dst.File
	err    error

	original string
}

func NewSourceFile(name string) *SourceFile {
	sf := &SourceFile{
		name: name,
	}

	sf.fset = token.NewFileSet()
	sf.dec = decorator.NewDecorator(sf.fset)

	return sf
}

func (sf *SourceFile) Decorator() *decorator.Decorator { return sf.dec }
func (sf *SourceFile) Path() string                    { return sf.name }
func (sf *SourceFile) FileSet() *token.FileSet         { return sf.fset }
func (sf *SourceFile) OriginalText() string            { return sf.original }
func (sf *SourceFile) Root() Node                      { return sf.root }
func (sf *SourceFile) Error() error                    { return sf.err }

func (sf *SourceFile) Parse(filename string, sourceCode string) (result Node, err error) {
	parsed, err := decorator.ParseFile(sf.fset, filename, sourceCode, parser.ParseComments)

	sf.parsed = parsed
	sf.err = err

	if parsed == nil {
		return nil, err
	}

	node := convertNode(parsed, sf)

	if sf.root == nil {
		sf.original = sourceCode
		sf.root = node
	}

	return node, err
}

func (sf *SourceFile) ToCode(node Node) (string, error) {
	var buf bytes.Buffer

	f, ok := node.Ast().(*dst.File)

	if !ok {
		decl, ok := node.Ast().(dst.Decl)

		if !ok {
			return "", errors.New("node is not a file or decl")
		}

		obj := &dst.Object{}

		switch decl := decl.(type) {
		case *dst.FuncDecl:
			obj.Kind = dst.Fun
		case *dst.GenDecl:
			switch decl.Tok {
			case token.CONST:
				obj.Kind = dst.Con
			case token.TYPE:
				obj.Kind = dst.Typ
			case token.VAR:
				obj.Kind = dst.Var
			}
		}

		f = &dst.File{
			Name:       sf.parsed.Name,
			Decls:      []dst.Decl{decl},
			Scope:      dst.NewScope(nil),
			Imports:    sf.parsed.Imports,
			Unresolved: sf.parsed.Unresolved,
			Decs:       sf.parsed.Decs,
		}

		f.Scope.Insert(&dst.Object{
			Kind: dst.Var,
			Decl: decl,
		})
	}

	err := decorator.Fprint(&buf, f)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
