package mdlang

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/dave/dst/decorator"
	"github.com/gomarkdown/markdown/ast"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

type SourceFile struct {
	psi.NodeBase

	name   string
	handle repofs.FileHandle

	l   *Language
	dec *decorator.Decorator

	root   psi.Node
	parsed ast.Node
	err    error

	original string
}

func NewSourceFile(l *Language, name string, handle repofs.FileHandle) *SourceFile {
	sf := &SourceFile{
		l: l,

		name:   name,
		handle: handle,
	}

	sf.Init(sf, sf.name)

	return sf
}

func (sf *SourceFile) Name() string                    { return sf.name }
func (sf *SourceFile) Language() psi.Language          { return sf.l }
func (sf *SourceFile) Decorator() *decorator.Decorator { return sf.dec }
func (sf *SourceFile) Path() string                    { return sf.name }
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
	sf.original = string(data)

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

func (sf *SourceFile) SetRoot(node ast.Node) error {
	sf.parsed = node
	sf.root = AstToPsi(sf.parsed)
	sf.root.SetParent(sf)

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

	parsed := mdutils.ParseMarkdown([]byte(sourceCode))

	if sf.root == nil {
		if err := sf.SetRoot(parsed); err != nil {
			return nil, err
		}

		return sf.root, nil
	}

	return AstToPsi(parsed), nil
}

func (sf *SourceFile) ToCode(node psi.Node) (string, error) {
	txt := string(mdutils.FormatMarkdown(node.(Node).Ast()))
	txt = strings.TrimSpace(txt)
	txt = strings.TrimRight(txt, "\n")
	return txt, nil
}

func (sf *SourceFile) MergeCompletionResults(ctx context.Context, scope psi.Scope, cursor psi.Cursor, newAst psi.Node) error {
	if newHeadings, ok := newAst.(*ast.Headings); ok {
		if currentHeadings, ok := cursor.Node().(*ast.Headings); ok {
			// Merge headings by name
			currentHeadings.Name += newHeadings.Name
			cursor.SetNode(currentHeadings)
			return nil
		}
	}

	cursor.Replace(newAst)

	return nil
}
