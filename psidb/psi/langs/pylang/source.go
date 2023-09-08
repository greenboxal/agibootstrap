package pylang

import (
	"bytes"
	"context"
	"fmt"
	"go/token"
	"io"

	"github.com/antlr4-go/antlr/v4"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
	"github.com/greenboxal/agibootstrap/psidb/psi/langs/pylang/pyparser"

	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering"
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

	sf.Init(sf)

	return sf
}

func (sf *SourceFile) Name() string               { return sf.name }
func (sf *SourceFile) Language() project.Language { return sf.l }
func (sf *SourceFile) Path() string               { return sf.name }
func (sf *SourceFile) OriginalText() string       { return sf.original }
func (sf *SourceFile) Root() psi.Node             { return sf.root }
func (sf *SourceFile) Error() error               { return sf.err }

func (sf *SourceFile) Load(ctx context.Context) error {
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

	_, err = sf.Parse(ctx, sf.name, string(data))

	sf.err = err

	return err
}

func (sf *SourceFile) Replace(ctx context.Context, code string) error {
	if code == sf.original {
		return nil
	}

	err := sf.handle.Put(bytes.NewBufferString(code))

	if err != nil {
		return err
	}

	return sf.Load(ctx)
}

func (sf *SourceFile) SetRoot(ctx context.Context, node antlr.ParserRuleContext, original string, tokens *antlr.CommonTokenStream) error {
	sf.original = original
	sf.parsed = node
	sf.tokens = tokens
	sf.rewriter = antlr.NewTokenStreamRewriter(tokens)

	sf.file = sf.l.project.FileSet().AddFile(sf.name, -1, len(sf.original))
	sf.file.SetLinesForContent([]byte(original))

	sf.root = AstToPsi(sf, sf.parsed)
	sf.root.SetParent(sf)

	return sf.Update(ctx)
}

func (sf *SourceFile) Parse(ctx context.Context, filename string, sourceCode string) (result psi.Node, err error) {
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
		if err := sf.SetRoot(ctx, parsed, sourceCode, tokens); err != nil {
			return nil, err
		}

		return sf.root, nil
	}

	return AstToPsi(sf, parsed), nil
}

type CharTokenizer struct {
}

func (c CharTokenizer) GetTokens(bytes []byte) (tokenizers.Tokens, error) {
	//TODO implement me
	panic("implement me")
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
		Write: func(w *rendering.TokenBuffer, node psi.Node) (err error) {
			n := node.(Node)

			hidden := sf.tokens.GetHiddenTokensToLeft(n.Tree().GetSourceInterval().Start, 1)

			for _, tk := range hidden {
				_, err := w.Write([]byte(tk.GetText()))

				if err != nil {
					return err
				}
			}

			if n.IsContainer() {
				for _, child := range n.Children() {
					_, err := w.WriteNode(renderer, child)

					if err != nil {
						return err
					}
				}
			} else {
				t := n.Token()
				str := t.GetText()

				_, err := w.Write([]byte(str))

				if err != nil {
					return err
				}
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

func (sf *SourceFile) MergeCompletionResults(ctx context.Context, scope project.Scope, cursor psi.Cursor, newSource project.SourceFile, newAst psi.Node) error {

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
