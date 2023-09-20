package jukebox

import (
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
	"github.com/go-errors/errors"

	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

type Lexer struct {
	sparsing.LexerStream
}

func NewLexer(handler ...sparsing.LexerTokenHandler) *Lexer {
	l := &Lexer{}

	l.LexerStream = sparsing.NewLexerStreamWithHandlerStack(l, handler)

	return l
}

func (l *Lexer) ConsumeStream(ctx sparsing.StreamingLexerContext) error {
	for ctx.Remaining() > 0 {
		switch ctx.CurrentToken().Kind {
		case sparsing.TokenKindInvalid:
			ctx.PushStart(TokenKindLine, string(ctx.NextChar()))
			continue
		case TokenKindLine:
			ch := ctx.NextChar()

			switch ch {
			case sparsing.RuneEOS:
				return nil
			case sparsing.RuneEOF:
				ctx.PushEnd()
			case '\n':
				ctx.PushEnd()
			default:
				ctx.AppendValue(string(ch))
			}
		default:
			return sparsing.ErrInvalidToken
		}
	}

	return nil
}

type Parser struct {
	sparsing.ParserStream
}

func NewParser() *Parser {
	p := &Parser{}

	p.ParserStream = sparsing.NewParserStream()
	p.PushLexerHandler(&Lexer{LexerStream: p})
	p.PushTokenParser(p)

	return p
}

type CommandSheetLine struct {
	sparsing.CompositeNodeBase

	Commands []*Command
}

type CommandSheetNode struct {
	sparsing.CompositeNodeBase

	Lines []*CommandSheetLine
}

func (p *Parser) acceptAny(ctx sparsing.StreamingParserContext) error {
	ctx.BeginComposite(&CommandSheetNode{}, ctx.PeekToken(0))

	return nil
}

func (p *Parser) acceptCommandSheetLine(ctx sparsing.StreamingParserContext) error {
	current := ctx.CurrentNode().(*CommandSheetLine)
	tk := ctx.ConsumeNextToken(sparsing.TokenKindEOF, sparsing.TokenKindEOS, TokenKindLine)

	if tk.GetKind() == sparsing.TokenKindEOS {
		return nil
	}

	if tk.GetKind() == TokenKindLine && strings.TrimSpace(tk.GetValue()) != "" {
		lex, err := CommandSheetLexer.LexString("", tk.GetValue())

		if err != nil {
			return err
		}

		stream, err := lexer.Upgrade(lex, CommandSheetLexer.Symbols()["whitespace"])

		if err != nil {
			return err
		}

		node, err := CommandSheetParser.ParseFromLexer(stream)

		if err != nil {
			return err
		}

		current.Commands = append(current.Commands, node.Commands...)
	}

	ctx.EndComposite(tk)

	return nil
}

func (p *Parser) acceptCommandSheet(ctx sparsing.StreamingParserContext) error {
	current := ctx.CurrentNode().(*CommandSheetNode)
	tk := ctx.PeekToken(0)

	for ctx.ArgumentStackLength() > 0 {
		node := ctx.PopArgument()

		switch node := node.(type) {
		case *CommandSheetNode:
			current.Lines = append(current.Lines, node.Lines...)

		case *CommandSheetLine:
			current.Lines = append(current.Lines, node)

		default:
			return errors.New(sparsing.ErrInvalidToken)
		}
	}

	switch tk.GetKind() {
	case sparsing.TokenKindEOS:
		return nil
	case sparsing.TokenKindEOF:
		ctx.EndComposite(ctx.NextToken())
	case TokenKindLine:
		ctx.BeginComposite(&CommandSheetLine{}, tk)
	default:
		return sparsing.ErrInvalidToken
	}

	return nil
}

func (p *Parser) ConsumeTokenStream(ctx sparsing.StreamingParserContext) error {
	for ctx.RemainingTokens() > 0 {
		switch ctx.CurrentNode().(type) {
		case nil:
			if err := p.acceptAny(ctx); err != nil {
				return err
			}

		case *CommandSheetNode:
			if err := p.acceptCommandSheet(ctx); err != nil {
				return err
			}

		case *CommandSheetLine:
			if err := p.acceptCommandSheetLine(ctx); err != nil {
				return err
			}

		default:
			return errors.New(sparsing.ErrInvalidToken)
		}
	}

	return nil
}
