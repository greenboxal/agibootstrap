package combinator

import (
	"github.com/go-errors/errors"

	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing/jsonstream"
)

type BasicNode interface {
	IsTerminal() bool
}

type ParserReference struct {
	LexerNode
}

func (r ParserReference) Build() LexerNode {
	return r.LexerNode
}

func (r ParserReference) Bind(parser LexerNode) LexerNode {
	r.LexerNode = parser
	return r
}

var valueReference = &ParserReference{}

var String = NewTerminal(jsonstream.TokenKindString, `"([^"]|\\")*"`)
var Number = NewTerminal(jsonstream.TokenKindNumber, `-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?`)
var Comma = NewTerminal(jsonstream.TokenKindComma, `,`)
var Colon = NewTerminal(jsonstream.TokenKindColon, `:`)
var LBrace = NewTerminal(jsonstream.TokenKindOpenObject, `{`)
var RBrace = NewTerminal(jsonstream.TokenKindCloseObject, `}`)
var LBracket = NewTerminal(jsonstream.TokenKindOpenArray, `\[`)
var RBracket = NewTerminal(jsonstream.TokenKindCloseArray, `\]`)
var True = NewTerminal(jsonstream.TokenKindIdent, `true`)
var False = NewTerminal(jsonstream.TokenKindIdent, `false`)
var Boolean = NewChoiceParser(True, False)

var Null = NewTerminal(jsonstream.TokenKindIdent, `null`)

var Pair = NewSequenceParser(String, Colon, valueReference)
var Object = NewSequenceParser(
	LBrace,
	NewRepeatParser(NewSequenceParser(
		Pair,
		NewOptionalParser(Comma),
	), 0, -1),
	RBrace,
)
var Array = NewSequenceParser(
	LBracket,
	NewRepeatParser(NewSequenceParser(
		valueReference,
		NewOptionalParser(Comma),
	), 0, -1),
	RBracket,
)

var Value = valueReference.Bind(NewChoiceParser(
	String,
	Number,
	Boolean,
	Object,
	Array,
	Null,
))

func CollectTerminals(dst map[sparsing.TokenKind]*TerminalParser, node LexerNode) {
	var walk func(node LexerNode)

	seen := make(map[LexerNode]struct{})

	walk = func(node LexerNode) {
		if _, ok := seen[node]; ok {
			return
		}

		seen[node] = struct{}{}

		if tp, ok := node.(*TerminalParser); ok {
			dst[tp.kind] = tp
		} else {
			switch node := node.(type) {
			case *SequenceParser:
				for _, node := range node.nodes {
					CollectTerminals(dst, node)
				}
			case *ChoiceParser:
				for _, node := range node.nodes {
					CollectTerminals(dst, node)
				}
			case *RepeatParser:
				CollectTerminals(dst, node.node)
			}
		}
	}

	walk(node)
}

func NewParser(root LexerNode) sparsing.ParserStream {
	terminals := make(map[sparsing.TokenKind]*TerminalParser)

	CollectTerminals(terminals, root)

	p := sparsing.NewParserStream()

	p.PushLexerHandler(sparsing.StreamingLexerHandlerFunc(func(ctx sparsing.StreamingLexerContext) error {
		for ctx.Remaining() > 0 {
			for _, tp := range terminals {
				if tp.PeekMatch(ctx) {
					if err := tp.Match(ctx); err != nil {
						return err
					}

					continue
				}
			}

			return errors.Errorf("unexpected token: %s", ctx.PeekBuffer())
		}

		return nil
	}))

	p.PushTokenParser(sparsing.ParserTokenHandlerFunc(func(ctx sparsing.StreamingParserContext) error {
		for ctx.RemainingTokens() > 0 {
			current := ctx.CurrentNode()

			if current == nil {
				if err := root.Accept(ctx); err != nil {
					return err
				}
			} else {

			}
		}

		return nil
	}))

	return p
}
