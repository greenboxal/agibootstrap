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

func Parse(filename string, sourceCode string) *SourceFile {
	sf := NewSourceFile(filename)

	_, _ = sf.Parse(filename, sourceCode)

	return sf
}

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

func (sf *SourceFile) FindNodeByPos(pos token.Position) Node {
	var findNode func(n Node) Node
	findNode = func(n Node) Node {
		startPos := sf.dec.Decorations().StartOf(n.Node()).Pos()
		endPos := sf.dec.Decorations().EndOf(n.Node()).Pos()
		// check if pos is inside node range
		if startPos <= pos && endPos >= pos {
			// check for node with children
			// if matches, recurses on children
			if nn, ok := n.(NodeContainer); ok {
				for _, child := range nn.Children() {
					if res := findNode(child); res != nil {
						return res
					}
				}
			}
			return n
		}
		return nil
	}
	return findNode(sf.root)
}
