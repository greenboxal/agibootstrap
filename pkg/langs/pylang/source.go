package pylang

import (
	"bytes"
	"context"
	"go/token"
	"io"
	"strings"

	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/py"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"

	"github.com/go-python/gpython/parser"
)

type SourceFile struct {
	psi.NodeBase

	name   string
	handle repofs.FileHandle

	l *Language

	root   psi.Node
	parsed ast.Ast
	err    error

	original    string
	lines       []string
	lineOffsets []int
	file        *token.File
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

func (sf *SourceFile) Name() string           { return sf.name }
func (sf *SourceFile) Language() psi.Language { return sf.l }
func (sf *SourceFile) Path() string           { return sf.name }
func (sf *SourceFile) OriginalText() string   { return sf.original }
func (sf *SourceFile) Root() psi.Node         { return sf.root }
func (sf *SourceFile) Error() error           { return sf.err }

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
	sf.original = ""
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

func (sf *SourceFile) SetRoot(node ast.Mod, original string) error {
	sf.original = original
	sf.parsed = node
	sf.root = AstToPsi(sf.parsed)
	sf.root.SetParent(sf)
	sf.lines = strings.Split(sf.original, "\n")
	sf.lineOffsets = make([]int, len(sf.lines))

	sf.file = sf.l.project.FileSet().AddFile(sf.name, -1, len(original))

	for i := range sf.lineOffsets {
		if i == 0 {
			sf.lineOffsets[i] = 0
		} else {
			sf.lineOffsets[i] = sf.lineOffsets[i-1] + len(sf.lines[i-1]) + 1
		}

		sf.file.AddLine(sf.lineOffsets[i])
	}

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

	reader := bytes.NewBufferString(sourceCode)
	parsed, err := parser.Parse(reader, filename, py.ExecMode)

	if err != nil {
		sf.err = err

		return nil, err
	}

	if sf.root == nil {
		if err := sf.SetRoot(parsed, sourceCode); err != nil {
			return nil, err
		}

		return sf.root, nil
	}

	return AstToPsi(parsed), nil
}

func (sf *SourceFile) ToCode(node psi.Node) (mdutils.CodeBlock, error) {
	n := node.(Node)
	ns := n.NextSibling()

	initialLine := n.Ast().GetLineno()
	initialCol := n.Ast().GetColOffset()
	lastLine := -1
	lastCol := -1

	if ns != nil {
		ns := ns.(Node)

		lastLine = ns.Ast().GetLineno()
		lastCol = ns.Ast().GetColOffset()
	}

	txt := sf.getRange(initialLine, initialCol, lastLine, lastCol)

	return mdutils.CodeBlock{
		Language: string(LanguageID),
		Code:     txt,
		Filename: sf.Name(),
	}, nil
}

func (sf *SourceFile) MergeCompletionResults(ctx context.Context, scope psi.Scope, cursor psi.Cursor, newAst psi.Node) error {
	cursor.Replace(newAst)

	return nil
}

func (sf *SourceFile) getRange(line int, col int, line2 int, col2 int) string {
	start := sf.file.Pos(sf.lineOffsets[line] + col)
	startOffset := sf.file.Offset(start)

	if line2 == -1 && col2 == -1 {
		return sf.original[startOffset:]
	}

	end := sf.file.Pos(sf.lineOffsets[line2] + col2)
	endOffset := sf.file.Offset(end)

	return sf.original[startOffset:endOffset]
}

type ByteTokenizer struct{}

func (b *ByteTokenizer) Count(text string) (int, error) {
	return len(text), nil
}

func (b *ByteTokenizer) Encode(text string) ([]byte, error) {
	return []byte(text), nil
}

func (b *ByteTokenizer) Decode(tokens []byte) (string, error) {
	return string(tokens), nil
}
