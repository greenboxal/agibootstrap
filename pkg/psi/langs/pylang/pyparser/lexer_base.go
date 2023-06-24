package pyparser

import (
	"github.com/antlr4-go/antlr/v4"
)

type Python3LexerBase struct {
	*antlr.BaseLexer

	opened    int
	lastToken antlr.Token
	indents   []int
	buffer    []antlr.Token
}

const TabSize = 8

func (l *Python3LexerBase) EmitToken(token antlr.Token) {
	l.BaseLexer.EmitToken(token)

	l.buffer = append(l.buffer, token)
	l.lastToken = token
}

func (l *Python3LexerBase) NextToken() antlr.Token {
	input := l.GetInputStream()

	// Check if the end-of-file is ahead and there are still some DEDENTS expected.
	if input.LA(1) == -1 && len(l.indents) > 0 {
		if len(l.buffer) == 0 || l.buffer[len(l.buffer)].GetTokenType() != Python3LexerLINE_BREAK {
			// First emit an extra line break that serves as the end of the statement.
			l.emit(Python3LexerLINE_BREAK, antlr.TokenDefaultChannel, "")
		}

		// Now emit as much DEDENT tokens as needed.
		for len(l.indents) != 0 {
			l.emit(Python3LexerDEDENT, antlr.TokenDefaultChannel, "")
			l.indents = l.indents[:len(l.indents)-1]
		}
	}

	if len(l.buffer) == 0 {
		return l.BaseLexer.NextToken()
	}

	result := l.buffer[0]
	l.buffer = l.buffer[1:]

	return result
}

func (l *Python3LexerBase) IncIndentLevel() {
	l.opened++
}

func (l *Python3LexerBase) DecIndentLevel() {
	if l.opened > 0 {
		l.opened--
	}
}

func (l *Python3LexerBase) HandleNewLine() {
	l.emit(Python3LexerNEWLINE, antlr.TokenHiddenChannel, l.GetText())

	input := l.GetInputStream()
	next := input.LA(1)

	// Process whitespaces in HandleSpaces
	if next != ' ' && next != '\t' && l.IsNotNewLineOrComment(next) {
		l.ProcessNewLine(0)
	}
}

func (l *Python3LexerBase) HandleSpaces() {
	input := l.GetInputStream()
	next := input.LA(1)

	if (l.lastToken == nil || l.lastToken.GetTokenType() == Python3LexerNEWLINE) && l.IsNotNewLineOrComment(next) {
		// Calculates the indentation of the provided spaces, taking the
		// following rules into account:
		//
		// "Tabs are replaced (from left to right) by one to eight spaces
		//  such that the total number of characters up to and including
		//  the replacement is a multiple of eight [...]"
		//
		//  -- https://docs.python.org/3.1/reference/lexical_analysis.html#indentation

		indent := 0
		text := l.GetText()

		for i := 0; i < len(text); i++ {
			if text[i] == '\t' {
				indent += TabSize - indent%TabSize
			} else {
				indent++
			}
		}

		l.ProcessNewLine(indent)
	}

	l.emit(Python3LexerWS, antlr.TokenHiddenChannel, l.GetText())
}

func (l *Python3LexerBase) emit(tokenType, channel int, text string) {
	charIndex := l.GetCharIndex()
	token := antlr.NewCommonToken(l.GetTokenSourceCharStreamPair(), tokenType, channel, charIndex-len(text), charIndex-1)
	token.SetText(text)

	l.EmitToken(token)
}

func (l *Python3LexerBase) IsNotNewLineOrComment(next int) bool {
	return l.opened == 0 && next != '\r' && next != '\n' && next != '\f' && next != '#'
}

func (l *Python3LexerBase) ProcessNewLine(indent int) {
	l.emit(Python3LexerLINE_BREAK, antlr.TokenDefaultChannel, "")

	previous := 0

	if len(l.indents) > 0 {
		previous = l.indents[len(l.indents)-1]
	}

	if indent > previous {
		l.indents = append(l.indents, indent)
		l.emit(Python3LexerINDENT, antlr.TokenDefaultChannel, "")
	} else {
		// Possibly emit more than 1 DEDENT token.
		for len(l.indents) != 0 && l.indents[len(l.indents)-1] > indent {
			l.emit(Python3LexerDEDENT, antlr.TokenDefaultChannel, "")
			l.indents = l.indents[:len(l.indents)-1]
		}
	}
}
