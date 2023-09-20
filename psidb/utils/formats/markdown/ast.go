package markdown

import (
	"unicode"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"

	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing/jsonstream"
)

type Container struct {
	sparsing.CompositeNodeBase

	Children []sparsing.Node
}

func (c *Container) ConsumeTokenStream(ctx sparsing.StreamingParserContext) error {
	for ctx.ArgumentStackLength() > 0 {
		c.Children = append(c.Children, ctx.PopArgument())
	}

	for ctx.RemainingTokens() > 0 {
		tk := ctx.PeekToken(0)

		switch tk.GetKind() {
		case TokenKindEOF:
			return nil
		case TokenKindEOS:
			return nil
		case TokenKindText:
			ctx.BeginComposite(&Text{}, nil)
			return nil
		case TokenKindLineBreak:
			ctx.NextToken()
		case TokenKindSymbol:
			switch tk.GetValue()[0] {
			case '#':
				ctx.BeginComposite(&Heading{}, nil)
				return nil
			case '`':
				if tk.GetValue() == "```" {
					ctx.BeginComposite(&CodeBlock{}, nil)
					return nil
				} else {
					ctx.BeginComposite(&Text{}, nil)
					return nil
				}
			default:
				ctx.BeginComposite(&Text{}, nil)
				return nil
			}
		}
	}

	return nil
}

func (c *Container) ConsumeStream(ctx sparsing.StreamingLexerContext) error {
	for ctx.Remaining() > 0 {
		switch ctx.CurrentToken().Kind {
		case sparsing.TokenKindInvalid:
			if err := c.consumeAny(ctx); err != nil {
				return err
			}

		case TokenKindLineBreak:
			ch := ctx.LA(0)

			if ch == '\n' {
				ctx.AppendValue(string(ctx.NextChar()))
			} else {
				ctx.PushEnd()
			}

		case TokenKindSymbol:
			ch := ctx.LA(0)

			if ch == rune(ctx.CurrentToken().Value[0]) {
				ctx.AppendValue(string(ctx.NextChar()))
			} else {
				ctx.PushEnd()
			}

		case TokenKindSpace:
			ch := ctx.LA(0)

			if ch == rune(ctx.CurrentToken().Value[0]) {
				ctx.AppendValue(string(ctx.NextChar()))
			} else {
				ctx.PushEnd()
			}

		case TokenKindText:
			ch := ctx.LA(0)

			if !(unicode.IsSpace(ch) || unicode.IsPunct(ch) || unicode.IsSymbol(ch) || unicode.IsMark(ch)) {
				ctx.AppendValue(string(ctx.NextChar()))
			} else {
				ctx.PushEnd()
			}
		}
	}

	return nil
}

func (c *Container) consumeAny(ctx sparsing.StreamingLexerContext) error {
	ch := ctx.NextChar()

	switch ch {
	case sparsing.RuneEOF:
		ctx.PushEnd()
	case '\n':
		ctx.PushSingle(TokenKindLineBreak, "\n")
	default:
		if unicode.IsSpace(ch) {
			ctx.PushStart(TokenKindSpace, string(ch))
		} else if unicode.IsPunct(ch) || unicode.IsSymbol(ch) || unicode.IsMark(ch) {
			ctx.PushStart(TokenKindSymbol, string(ch))
		} else {
			ctx.PushStart(TokenKindText, string(ch))
		}
	}

	return nil
}

type Text struct {
	sparsing.CompositeNodeBase

	Content string   `json:"content"`
	Parsed  ast.Node `json:"parsed"`
}

func (t *Text) ConsumeTokenStream(ctx sparsing.StreamingParserContext) error {
	for ctx.RemainingTokens() > 0 {
		tk := ctx.PeekToken(0)

		switch tk.GetKind() {
		case TokenKindText:
			t.Content += ctx.NextToken().GetValue()
		case TokenKindSpace:
			t.Content += ctx.NextToken().GetValue()
		case TokenKindSymbol:
			switch tk.GetValue() {
			default:
				t.Content += ctx.NextToken().GetValue()
			}

		default:
			t.Parsed = markdown.Parse([]byte(t.Content), nil)

			ctx.EndComposite(nil)
			return nil
		}
	}

	return nil
}

type Heading struct {
	sparsing.CompositeNodeBase

	Marker sparsing.IToken
	Level  int
	Title  string
}

func (h *Heading) ConsumeTokenStream(ctx sparsing.StreamingParserContext) error {
	for ctx.RemainingTokens() > 0 {
		tk := ctx.PeekToken(0)

		switch tk.GetKind() {
		case TokenKindSymbol:
			if h.Marker == nil && tk.GetValue()[0] == '#' {
				h.Marker = ctx.NextToken()
				h.Level = len(h.Marker.GetValue())
			}
		case TokenKindText:
			h.Title += ctx.NextToken().GetValue()
		case TokenKindSpace:
			h.Title += ctx.NextToken().GetValue()
		case TokenKindLineBreak:
			fallthrough
		case TokenKindEOF:
			ctx.EndComposite(nil)
			return nil
		}
	}

	return nil
}

type Code struct {
	sparsing.CompositeNodeBase

	StartMarker sparsing.IToken
	Content     string
	EndMarker   sparsing.IToken
}

func (c *Code) ConsumeTokenStream(ctx sparsing.StreamingParserContext) error {
	for ctx.RemainingTokens() > 0 {
		tk := ctx.PeekToken(0)

		switch tk.GetKind() {
		case TokenKindSymbol:
			if c.StartMarker == nil && tk.GetValue() == "`" {
				c.StartMarker = ctx.NextToken()
			} else if c.EndMarker == nil && tk.GetValue() == "`" {
				c.EndMarker = ctx.NextToken()
				ctx.EndComposite(c.EndMarker)
			} else {
				c.Content += ctx.NextToken().GetValue()
			}
		case TokenKindText:
			fallthrough
		case TokenKindSpace:
			c.Content += ctx.NextToken().GetValue()
		case TokenKindLineBreak:
			fallthrough
		case TokenKindEOF:
			ctx.EndComposite(nil)
			return nil
		}
	}

	return nil
}

type CodeBlock struct {
	sparsing.CompositeNodeBase

	StartMarker sparsing.IToken
	Language    string
	LF          sparsing.IToken
	Content     string
	EndMarker   sparsing.IToken
}

func (cb *CodeBlock) ConsumeTokenStream(ctx sparsing.StreamingParserContext) error {
	for ctx.RemainingTokens() > 0 {
		tk := ctx.PeekToken(0)

		switch tk.GetKind() {
		case TokenKindEOF:
			ctx.EndComposite(nil)
			return nil
		case TokenKindSymbol:
			if cb.StartMarker == nil && tk.GetValue() == "```" {
				cb.StartMarker = ctx.NextToken()
			} else if cb.EndMarker == nil && tk.GetValue() == "```" {
				cb.EndMarker = ctx.NextToken()
				ctx.EndComposite(cb.EndMarker)
			} else if cb.LF != nil {
				cb.Content += ctx.NextToken().GetValue()
			} else {
				cb.Language += ctx.NextToken().GetValue()
			}
		case TokenKindText:
			fallthrough
		case TokenKindSpace:
			if cb.LF != nil {
				cb.Content += ctx.NextToken().GetValue()
			} else {
				cb.Language += ctx.NextToken().GetValue()
			}
		case TokenKindLineBreak:
			if cb.LF == nil {
				cb.LF = ctx.NextToken()

				if cb.Language == "json" {
					ctx.BeginComposite(&jsonstream.Object{}, nil)
				}
			} else {
				cb.Content += ctx.NextToken().GetValue()
			}

			return nil
		}
	}

	return nil
}
