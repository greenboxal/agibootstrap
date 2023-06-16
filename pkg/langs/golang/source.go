package golang

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

type SourceFile struct {
	name   string
	handle repofs.FileHandle

	dec  *decorator.Decorator
	fset *token.FileSet

	root   psi.Node
	parsed *dst.File
	err    error

	original string
}

func NewSourceFile(fset *token.FileSet, name string, handle repofs.FileHandle) *SourceFile {
	sf := &SourceFile{
		name:   name,
		fset:   fset,
		handle: handle,
	}

	sf.dec = decorator.NewDecorator(sf.fset)

	return sf
}

func (sf *SourceFile) Decorator() *decorator.Decorator { return sf.dec }
func (sf *SourceFile) Path() string                    { return sf.name }
func (sf *SourceFile) FileSet() *token.FileSet         { return sf.fset }
func (sf *SourceFile) OriginalText() string            { return sf.original }
func (sf *SourceFile) Root() psi.Node                  { return sf.root }
func (sf *SourceFile) Error() error                    { return sf.err }

func (sf *SourceFile) Load() error {
	file, err := sf.handle.Get()

	if err != nil {
		return err
	}

	data, err := io.ReadAll(file)

	if err != nil {
		return err
	}

	sf.root = nil
	sf.parsed = nil
	sf.err = nil

	_, err = sf.Parse(sf.name, string(data))

	sf.err = err

	return err
}

func (sf *SourceFile) Replace(code string) error {
	if code == sf.original {
		return nil
	}

	err := sf.handle.Put(bytes.NewBufferString(code))

	if err != nil {
		return err
	}

	return sf.Load()
}

func (sf *SourceFile) SetRoot(node *ast.File) error {
	parsed, err := decorator.Decorate(sf.fset, node)

	if err != nil {
		return err
	}

	sf.parsed = parsed.(*dst.File)
	sf.root = AstToPsi(sf.parsed)

	return nil
}

func (sf *SourceFile) Parse(filename string, sourceCode string) (result psi.Node, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = r.(error)
			}

			err = errors.Wrap(err, "panic while parsing file: "+filename)
		}
	}()

	parsed, err := decorator.ParseFile(sf.fset, filename, sourceCode, parser.ParseComments)

	sf.parsed = parsed
	sf.err = err

	if parsed == nil {
		return nil, err
	}

	node := AstToPsi(parsed)

	if sf.root == nil {
		sf.original = sourceCode
		sf.root = node
	}

	return node, err
}

func (sf *SourceFile) ToCode(node psi.Node) (string, error) {
	var buf bytes.Buffer

	n := node.(Node).Ast()
	f, ok := n.(*dst.File)

	if !ok {
		var decls []dst.Decl

		obj := &dst.Object{}

		switch n := n.(type) {
		case *dst.FuncDecl:
			obj.Kind = dst.Fun

			decls = append(decls, n)

		case *dst.GenDecl:
			switch n.Tok {
			case token.CONST:
				obj.Kind = dst.Con
			case token.TYPE:
				obj.Kind = dst.Typ
			case token.VAR:
				obj.Kind = dst.Var
			}

			decls = append(decls, n)

		case *dst.TypeSpec:
			obj.Kind = dst.Typ

			decls = append(decls, &dst.GenDecl{
				Tok:   token.TYPE,
				Specs: []dst.Spec{n},
			})

		default:
			return "", errors.New("node is not a file or decl")
		}

		f = &dst.File{
			Name:       sf.parsed.Name,
			Decls:      decls,
			Scope:      dst.NewScope(nil),
			Imports:    sf.parsed.Imports,
			Unresolved: sf.parsed.Unresolved,
			Decs:       sf.parsed.Decs,
		}

		f.Scope.Insert(obj)
	}

	err := decorator.Fprint(&buf, f)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
