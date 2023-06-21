package pyparser

import (
	"regexp"

	"github.com/antlr4-go/antlr/v4"
)

type Python3LexerBase struct {
	*antlr.BaseLexer

	tokens    []antlr.Token
	indents   []int
	opened    int
	lastToken antlr.Token
	input     antlr.CharStream
}

func NewPython3LexerBase(input antlr.CharStream) *Python3LexerBase {
	return &Python3LexerBase{
		BaseLexer: antlr.NewBaseLexer(input),

		input:   input,
		tokens:  make([]antlr.Token, 0),
		indents: make([]int, 0),
	}
}

func (p *Python3LexerBase) EmitToken(t antlr.Token) {
	p.tokens = append(p.tokens, t)
}

func (p *Python3LexerBase) NextToken() antlr.Token {
	if p.input.LA(1) == -1 && len(p.indents) != 0 {
		// Remove any trailing EOF tokens from our buffer.
		for i := len(p.tokens) - 1; i >= 0; i-- {
			if p.tokens[i].GetTokenType() == -1 {
				p.tokens = p.tokens[:i]
			}
		}

		// First emit an extra line break that serves as the end of the statement.
		p.EmitToken(p.CommonToken(Python3LexerNEWLINE, "\n"))

		// Now emit as much DEDENT tokens as needed.
		for len(p.indents) > 0 {
			p.EmitToken(p.CreateDedent())
			p.indents = p.indents[:len(p.indents)-1]
		}

		// Put the EOF back on the token stream.
		p.EmitToken(p.CommonToken(-1, "<EOF>"))
	}

	next := p.BaseLexer.NextToken() // You will need to replace this part

	if next.GetChannel() == 0 {
		p.lastToken = next
	}

	if len(p.tokens) == 0 {
		return next
	} else {
		token := p.tokens[0]
		p.tokens = p.tokens[1:]
		return token
	}
}

func (p *Python3LexerBase) CreateDedent() antlr.Token {
	dedent := p.CommonToken(Python3LexerDEDENT, "")
	//dedent.SetLine(p.lastToken.GetLine())
	return dedent
}

func (p *Python3LexerBase) CommonToken(ttype int, text string) *antlr.CommonToken {
	stop := p.GetCharIndex() - 1 // You will need to replace this part
	start := stop
	if len(text) != 0 {
		start = stop - len(text) + 1
	}
	return antlr.NewCommonToken(p.GetTokenSourceCharStreamPair(), ttype, 0, start, stop) // You will need to replace this part
}

func GetIndentationCount(spaces string) int {
	count := 0
	for _, ch := range spaces {
		switch ch {
		case '\t':
			count += 8 - (count % 8)
		default:
			count++
		}
	}

	return count
}

func (p *Python3LexerBase) AtStartOfInput() bool {
	return p.GetCharPositionInLine() == 0 && p.GetLine() == 1 // You will need to replace this part
}

func (p *Python3LexerBase) OpenBrace() {
	p.opened++
}

func (p *Python3LexerBase) CloseBrace() {
	p.opened--
}

var newLineRegex = regexp.MustCompile("[^\\r\\n\\f]+")
var spaceRegex = regexp.MustCompile("[\\r\\n\\f]+")

func (p *Python3LexerBase) OnNewLine() {
	newLine := newLineRegex.ReplaceAllString(p.GetText(), "")
	spaces := spaceRegex.ReplaceAllString(p.GetText(), "")

	// ...
	// omitted for brevity
	// ...
	// Strip newlines inside open clauses except if we are near EOF. We keep NEWLINEs near EOF to
	// satisfy the final newline needed by the single_put rule used by the REPL.
	next := p.input.LA(1)
	nextnext := p.input.LA(2)
	if p.opened > 0 || (nextnext != -1 && (next == '\r' || next == '\n' || next == '\f' || next == '#')) {
		// If we're inside a list or on a blank line, ignore all indents,
		// dedents and line breaks.
		p.Skip()
	} else {
		p.EmitToken(p.CommonToken(Python3LexerNEWLINE, newLine))
		indent := GetIndentationCount(spaces)
		previous := 0
		if len(p.indents) != 0 {
			previous = p.indents[len(p.indents)-1]
		}

		if indent == previous {
			// skip indents of the same size as the present indent-size
			p.Skip()
		} else if indent > previous {
			p.indents = append(p.indents, indent)
			p.EmitToken(p.CommonToken(Python3LexerINDENT, spaces))
		} else {
			// Possibly emit more than 1 DEDENT token.
			for len(p.indents) != 0 && p.indents[len(p.indents)-1] > indent {
				p.EmitToken(p.CreateDedent())
				p.indents = p.indents[:len(p.indents)-1]
			}
		}
	}
}

func (p *Python3LexerBase) Reset() {
	p.tokens = make([]antlr.Token, 0)
	p.indents = make([]int, 0)
	p.opened = 0
	p.lastToken = nil
	p.BaseLexer.Reset()
}
