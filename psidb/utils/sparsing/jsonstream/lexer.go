package jsonstream

import (
	"fmt"
	"unicode"

	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

type CompositeNodeBase struct {
	sparsing.CompositeNodeBase
	NodeBase
}

type TerminalNodeBase struct {
	sparsing.TerminalNodeBase
	NodeBase
}

type NodeBase struct {
	path ParserPath
}

func (l *NodeBase) ConsumeStream(ctx sparsing.StreamingLexerContext) error {
	consumeAny := func() error {
		for {
			ch := ctx.NextChar()

			switch ch {
			case sparsing.RuneEOS:
				return nil

			case sparsing.RuneEOF:
				return nil

			case ' ', '\t', '\n', '\r':
				continue

			case '{':
				ctx.PushSingle(TokenKindOpenObject, string(ch))

			case '}':
				ctx.PushSingle(TokenKindCloseObject, string(ch))

			case '[':
				ctx.PushSingle(TokenKindOpenArray, string(ch))

			case ']':
				ctx.PushSingle(TokenKindCloseArray, string(ch))

			case ':':
				ctx.PushSingle(TokenKindColon, string(ch))

			case ',':
				ctx.PushSingle(TokenKindComma, string(ch))

			case '"':
				ctx.PushStart(TokenKindString, string(ch))

			default:
				if unicode.IsLetter(ch) || ch == '_' {
					ctx.PushStart(TokenKindIdent, string(ch))
				} else if unicode.IsDigit(ch) || ch == '-' || ch == '.' {
					ctx.PushStart(TokenKindNumber, string(ch))
				} else {
					return fmt.Errorf("unexpected character: %v", ch)
				}
			}

			return nil
		}
	}

	for ctx.Remaining() > 0 {
		switch ctx.CurrentToken().Kind {
		case TokenKindInvalid:
			if err := consumeAny(); err != nil {
				return err
			}

		case TokenKindIdent:
			ch := ctx.LA(0)

			if ch == sparsing.RuneEOS {
				return nil
			} else if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
				ctx.AppendValue(string(ctx.NextChar()))
			} else {
				ctx.PushEnd()
			}

		case TokenKindString:
			ch := ctx.NextChar()

			ctx.AppendValue(string(ch))

			if ch == sparsing.RuneEOS {
				return nil
			} else if ch == '"' {
				ctx.PushEnd()
			} else if ch == '\\' {
				ctx.AppendValue(string(ctx.NextChar()))
			}

		case TokenKindNumber:
			ch := ctx.LA(0)

			if ch == sparsing.TokenKindEOS {
				return nil
			} else if unicode.IsDigit(ch) || ch == '.' {
				ctx.AppendValue(string(ctx.NextChar()))
			} else {
				ctx.PushEnd()
			}

		default:
			return fmt.Errorf("unexpected token kind: %v", ctx.CurrentToken().Kind)
		}
	}

	return nil
}
