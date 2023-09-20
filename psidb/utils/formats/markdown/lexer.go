package markdown

import (
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

type Lexer struct {
}

func (l *Lexer) ConsumeStream(ctx sparsing.StreamingLexerContext) error {
	for ctx.Remaining() > 0 {
		switch ctx.CurrentToken().Kind {
		case sparsing.TokenKindInvalid:
			if err := l.consumeAny(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Lexer) consumeAny(ctx sparsing.StreamingLexerContext) error {

	return nil
}
