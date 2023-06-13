package psi

import (
	"bytes"
	"errors"
	"go/parser"
	"go/token"
	"html"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
)

type SourceFile struct {
	name   string
	dec    *decorator.Decorator
	parsed *dst.File
	err    error

	root *Container
}

func NewSourceFile(name string) *SourceFile {
	return &SourceFile{
		name: name,
		dec:  decorator.NewDecorator(token.NewFileSet()),
	}
}

func Parse(filename string, sourceCode string) (*SourceFile, error) {
	sf := NewSourceFile(filename)

	_, err := sf.Parse(filename, sourceCode)

	if err != nil {
		return nil, err
	}

	return sf, nil
}

func (sf *SourceFile) Path() string            { return sf.name }
func (sf *SourceFile) FileSet() *token.FileSet { return sf.dec.Fset }
func (sf *SourceFile) Root() *Container        { return sf.root }
func (sf *SourceFile) Error() error            { return sf.err }

func (sf *SourceFile) Parse(filename string, sourceCode string) (result Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = r.(error)
			}
		}
	}()

	containsQuote := strings.Contains(sourceCode, "&#34;")

	if containsQuote {
		sourceCode = html.UnescapeString(sourceCode)
	}

	parsed, err := sf.dec.ParseFile(filename, sourceCode, parser.ParseComments)

	sf.parsed = parsed
	sf.err = err

	if parsed == nil {
		return nil, err
	}

	node := convertNode(parsed, sf).(*Container)

	if sf.root == nil {
		sf.root = node
	}

	return node, err
}

func (sf *SourceFile) ToCode(node Node) (string, error) {
	var buf bytes.Buffer

	f, ok := node.Node().(*dst.File)

	if !ok {
		decl, ok := node.Node().(dst.Decl)

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
