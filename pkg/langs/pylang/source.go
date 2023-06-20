package pylang

import (
	"bytes"
	"context"
	"fmt"
	"go/token"
	"io"

	"github.com/antlr4-go/antlr/v4"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/langs/pylang/pyparser"
	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

type SourceFile struct {
	psi.NodeBase

	name   string
	handle repofs.FileHandle

	l *Language

	root   psi.Node
	parsed antlr.ParserRuleContext
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

func (sf *SourceFile) SetRoot(node antlr.ParserRuleContext, original string) error {
	sf.original = original
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

	reader := bytes.NewBufferString(sourceCode)
	stream := antlr.NewIoStream(reader)
	lexer := pyparser.NewPython3Lexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, 0)
	parser := pyparser.NewPython3Parser(tokenStream)

	if parser.HasError() {
		sf.err = fmt.Errorf("%s", parser.GetError())
	}

	parsed := parser.File_input()

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
	txt := n.Ast().GetText()

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
