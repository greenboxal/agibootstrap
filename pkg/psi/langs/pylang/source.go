package pylang

import (
	"bytes"
	"context"
	"fmt"
	"go/token"
	"io"

	"github.com/antlr4-go/antlr/v4"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/psi/langs/pylang/pyparser"

	"github.com/greenboxal/agibootstrap/pkg/platform/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
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

	stream := antlr.NewInputStream(sourceCode)
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

type CharTokenizer struct {
}

func (c CharTokenizer) Count(text string) (int, error) {
	return len(text), nil
}

func (sf *SourceFile) ToCode(node psi.Node) (mdutils.CodeBlock, error) {
	var renderer *rendering.PruningRenderer
	n := node.(Node)

	renderer = &rendering.PruningRenderer{
		Tokenizer: &CharTokenizer{},
		Weight: func(state *rendering.NodeState, node psi.Node) float32 {
			return 1
		},
		Write: func(w *rendering.TokenBuffer, node psi.Node) (total int, err error) {
			n := node.(Node)

			hidden := sf.tokens.GetHiddenTokensToLeft(n.Tree().GetSourceInterval().Start, 1)

			for _, tk := range hidden {
				n, err := w.Write([]byte(tk.GetText()))

				if err != nil {
					return total, err
				}

				total += n
			}

			if n.IsContainer() {
				for _, child := range n.Children() {
					num, err := w.WriteNode(renderer, child)

					if err != nil {
						return total, err
					}

					total += num
				}
			} else {
				t := n.Token()
				str := t.GetText()

				return w.Write([]byte(str))
			}

			return
		},
	}

	buf := bytes.NewBuffer(nil)
	_, err := renderer.Render(n, buf)

	if err != nil {
		panic(err)
	}

	txt := buf.String()
	txt = n.Tree().GetText()

	return mdutils.CodeBlock{
		Language: string(LanguageID),
		Code:     txt,
		Filename: sf.Name(),
	}, nil
}

func (sf *SourceFile) GetAdjustedTokenInterval(n Node) antlr.Interval {
	interval := n.Tree().GetSourceInterval()
	hiddenTokens := sf.tokens.GetHiddenTokensToLeft(interval.Start, 1)

	if len(hiddenTokens) > 0 {
		interval.Start = hiddenTokens[0].GetStart()
	}

	return interval
}

func (sf *SourceFile) MergeCompletionResults(ctx context.Context, scope psi.Scope, cursor psi.Cursor, newSource psi.SourceFile, newAst psi.Node) error {

	return nil
}

func (sf *SourceFile) getRange(txt string, start antlr.Token, end antlr.Token) string {
	startLinePos := sf.file.LineStart(start.GetLine())
	startLineOffset := sf.file.Offset(startLinePos) + start.GetColumn()

	if end == nil {
		return sf.original[startLineOffset:]
	}

	endLinePos := sf.file.LineStart(end.GetLine())
	endLineOffset := sf.file.Offset(endLinePos) + end.GetColumn()

	return txt[startLineOffset:endLineOffset]
}
