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

	original string
	file     *token.File
	tokens   *antlr.CommonTokenStream
	rewriter *antlr.TokenStreamRewriter
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

func (sf *SourceFile) SetRoot(node antlr.ParserRuleContext, original string, tokens *antlr.CommonTokenStream) error {
	sf.original = original
	sf.parsed = node
	sf.tokens = tokens
	sf.rewriter = antlr.NewTokenStreamRewriter(tokens)

	sf.file = sf.l.project.FileSet().AddFile(sf.name, -1, len(sf.original))
	sf.file.SetLinesForContent([]byte(original))

	sf.root = AstToPsi(sf, sf.parsed)
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
	tokens := antlr.NewCommonTokenStream(lexer, 0)
	parser := pyparser.NewPython3Parser(tokens)

	parsed := parser.File_input()

	if parser.HasError() {
		sf.err = fmt.Errorf("%s", parser.GetError())

		return nil, sf.err
	}

	if sf.root == nil {
		if err := sf.SetRoot(parsed, sourceCode, tokens); err != nil {
			return nil, err
		}

		return sf.root, nil
	}

	return AstToPsi(sf, parsed), nil
}

func (sf *SourceFile) ToCode(node psi.Node) (mdutils.CodeBlock, error) {
	n := node.(Node)

	var start, end antlr.Token

	start = n.Ast().GetStart()
	end = n.Ast().GetStop()

	hiddenStart := sf.tokens.GetHiddenTokensToLeft(start.GetTokenIndex(), 2)

	if len(hiddenStart) > 0 {
		start = hiddenStart[0]
	}

	ns := n.NextSibling()

	if ns != nil {
		if ns, ok := ns.(Node); ok {
			end = ns.Ast().GetStart()
		}
	} else {
		end = nil
	}

	rewriter := antlr.NewTokenStreamRewriter(sf.tokens)
	txt := rewriter.GetTextDefault()

	txt = sf.getRange(start, end)

	return mdutils.CodeBlock{
		Language: string(LanguageID),
		Code:     txt,
		Filename: sf.Name(),
	}, nil
}

func (sf *SourceFile) MergeCompletionResults(ctx context.Context, scope psi.Scope, cursor psi.Cursor, newSource psi.SourceFile, newAst psi.Node) error {
	cursor.Replace(newAst)

	return nil
}

func (sf *SourceFile) getRange(start antlr.Token, end antlr.Token) string {
	startLinePos := sf.file.LineStart(start.GetLine())
	startLineOffset := sf.file.Offset(startLinePos) + start.GetColumn()

	if end == nil {
		return sf.original[startLineOffset:]
	}

	endLinePos := sf.file.LineStart(end.GetLine())
	endLineOffset := sf.file.Offset(endLinePos) + end.GetColumn()

	return sf.original[startLineOffset:endLineOffset]
}
