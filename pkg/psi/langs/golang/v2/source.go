package golang

import (
	"context"
	"go/parser"
	"go/token"
	"io"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/analysis"
	"github.com/greenboxal/agibootstrap/pkg/psi/lex"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

func NewSourceFile(l *Language, name string, handle repofs.FileHandle) project.SourceFile {
	sf := &SourceFile{}
	sf.name = name
	sf.language = l
	sf.handle = handle

	sf.Init(sf)

	return sf
}

type SourceFile struct {
	project.SourceFileBase

	name     string
	language *Language

	root Node
	err  error

	buffer lex.StringTokenBuffer
	handle repofs.FileHandle

	fset *token.FileSet

	Contents string `json:"contents"`
}

var SourceFileType = psi.DefineNodeType[*SourceFile](psi.WithStubNodeType())

func (sf *SourceFile) Init(self psi.Node) {
	if sf.fset == nil {
		sf.fset = token.NewFileSet()
	}

	if sf.Contents != "" {
		sf.buffer.Write([]byte(sf.Contents))
		sf.fset.AddFile(sf.name, sf.fset.Base(), len(sf.Contents))
	}

	sf.SourceFileBase.Init(self, psi.WithNodeType(SourceFileType))

	analysis.DefineNodeScope(sf)
}

func (sf *SourceFile) Name() string               { return sf.name }
func (sf *SourceFile) Language() project.Language { return sf.language }
func (sf *SourceFile) Root() psi.Node             { return sf.root }
func (sf *SourceFile) Error() error               { return sf.err }
func (sf *SourceFile) OriginalText() string       { return sf.buffer.String() }

func (sf *SourceFile) Load(ctx context.Context) error {
	reader, err := sf.handle.Get()

	if err != nil {
		return err
	}

	data, err := io.ReadAll(reader)

	if err != nil {
		return err
	}

	if err := sf.buffer.Write(data); err != nil {
		return err
	}

	file, err := parser.ParseFile(sf.fset, sf.name, sf.buffer.ToSlice(nil), parser.ParseComments)

	if err != nil {
		return err
	}

	sf.root = GoAstToPsi(sf.fset, file)
	sf.root.SetParent(sf)
	sf.Contents = sf.buffer.String()

	analysis.DefineNodeScope(sf)

	if err := psi.Walk(sf.root, func(c psi.Cursor, entering bool) error {
		if !entering {
			return nil
		}

		n := c.Value()

		switch n := n.(type) {
		case *File:
			analysis.DefineNodeScope(n)

		case *FuncDecl:
			analysis.DefineNodeSymbol(n, n.GetName().Name)
			analysis.DefineNodeScope(n)

		case *BlockStmt:
			analysis.DefineNodeScope(n)

		case *GenDecl:
			for _, spec := range n.GetSpecs() {
				switch spec := spec.(type) {
				case *ValueSpec:
					for _, name := range spec.GetNames() {
						analysis.DefineNodeSymbol(n, name.Name)
					}
				case *TypeSpec:
					analysis.DefineNodeSymbol(n, spec.GetName().Name)
				case *ImportSpec:
					//TODO
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return sf.Update(ctx)
}

func (sf *SourceFile) Replace(ctx context.Context, code string) error {
	//TODO implement me
	panic("implement me")
}

func (sf *SourceFile) ToCode(node psi.Node) (mdutils.CodeBlock, error) {
	n := node.(Node)

	start := n.GoNodeBase().StartTokenPos
	end := n.GoNodeBase().EndTokenPos

	startPos := sf.fset.Position(start)
	endPos := sf.fset.Position(end)

	code := sf.buffer.StringSlice(startPos.Offset, endPos.Offset)

	return mdutils.CodeBlock{
		Language: "go",
		Code:     code,
	}, nil
}

func (sf *SourceFile) MergeCompletionResults(ctx context.Context, scope project.Scope, cursor psi.Cursor, newSource project.SourceFile, newAst psi.Node) error {
	//TODO implement me
	panic("implement me")
}
