package gensparse

import (
	"reflect"

	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

type TokenKind = sparsing.TokenKind
type Node = sparsing.Node
type CompositeNode = sparsing.CompositeNode
type TerminalNode = sparsing.TerminalNode
type Token = sparsing.IToken
type Position = sparsing.Position

type ParserNodeHandler[TToken Token, TNode Node] interface {
	ConsumeNode(ctx StreamingParserContext[TToken, TNode], node TNode) error
}

type ParserNodeHandlerFunc[TToken Token, TNode Node] func(ctx StreamingParserContext[TToken, TNode], node TNode) error

func (p ParserNodeHandlerFunc[TToken, TNode]) ConsumeNode(ctx StreamingParserContext[TToken, TNode], node TNode) error {
	return p(ctx, node)
}

type ParserTokenHandler[TToken Token, TNode Node] interface {
	ConsumeTokenStream(ctx StreamingParserContext[TToken, TNode]) error
}

type StreamingParserContext[TToken Token, TNode Node] interface {
	BeginComposite(s CompositeNode, start TToken)
	EndComposite(end TToken)
	PushTerminal(s TerminalNode, tk TToken)

	CurrentNode() TNode
	CurrentPath() []Node
	ArgumentStackLength() int
	PushArgument(arg ...TNode) ParserStream[TToken, TNode]
	PopArgument() TNode
	TryPopArgumentTo(ptr any) bool

	PushInStack(s TNode)
	PopToOutStack()

	RemainingTokens() int
	ConsumeNextToken(kinds ...TokenKind) TToken
	NextToken() TToken
	PeekToken(n int) sparsing.IToken

	PushTokenParser(handler ParserTokenHandler[TToken, TNode]) ParserStream[TToken, TNode]
	PopTokenParser() ParserStream[TToken, TNode]

	PushNodeConsumer(handler ParserNodeHandler[TToken, TNode]) ParserStream[TToken, TNode]
	PopNodeConsumer() ParserStream[TToken, TNode]
}

type ParserStream[TToken Token, TNode Node] interface {
	WriteTokens(tokens []TToken) (n int, err error)
	WriteToken(token TToken) error

	CurrentNode() TNode
	ArgumentStackLength() int
	PushArgument(arg ...TNode) ParserStream[TToken, TNode]
	PopArgument() TNode
	TryPopArgumentTo(ptr any) bool

	PushTokenParser(handler ParserTokenHandler[TToken, TNode]) ParserStream[TToken, TNode]
	PopTokenParser() ParserStream[TToken, TNode]

	PushNodeConsumer(handler ParserNodeHandler[TToken, TNode]) ParserStream[TToken, TNode]
	PopNodeConsumer() ParserStream[TToken, TNode]
}

type parserStream[TToken Token, TNode Node] struct {
	sparsing.LexerStream

	nodeHandlerStack  []ParserNodeHandler[TToken, TNode]
	tokenHandlerStack []ParserTokenHandler[TToken, TNode]

	buffer         []TToken
	bufferPosition int

	inStack  []TNode
	outStack []TNode

	err error

	closed bool
}

func NewParserStream[TToken Token, TNode Node]() ParserStream[TToken, TNode] {
	ps := &parserStream[TToken, TNode]{
		LexerStream: sparsing.NewLexerStream(),
	}

	ps.LexerStream.PushLexerConsumer(ps)

	return ps
}

func (ps *parserStream[TToken, TNode]) RemainingTokens() int {
	return len(ps.buffer) - ps.bufferPosition
}

func (ps *parserStream[TToken, TNode]) CurrentPath() []Node {
	p := make([]Node, len(ps.inStack))
	for i, n := range ps.inStack {
		p[i] = n
	}
	return p
}
func (ps *parserStream[TToken, TNode]) CurrentNode() (_ TNode) {
	if len(ps.inStack) == 0 {
		return
	}

	return ps.inStack[len(ps.inStack)-1]
}

func (ps *parserStream[TToken, TNode]) WriteTokens(tokens []TToken) (n int, err error) {
	if len(ps.tokenHandlerStack) == 0 {
		return 0, errors.New("no token handler set")
	}

	ps.buffer = append(ps.buffer, tokens...)

	return ps.advance()
}

func (ps *parserStream[TToken, TNode]) ConsumeToken(token sparsing.IToken) error {
	tk, ok := token.(TToken)

	if !ok {
		return errors.New("invalid token type")
	}

	return ps.WriteToken(tk)
}

func (ps *parserStream[TToken, TNode]) WriteToken(token TToken) error {
	_, err := ps.WriteTokens([]TToken{token})

	if err != nil {
		return err
	}

	return nil
}

func (ps *parserStream[TToken, TNode]) Close() error {
	if ps.closed {
		return nil
	}

	if err := ps.LexerStream.Close(); err != nil {
		return err
	}

	lastEnd := Position{}

	if len(ps.buffer) > 0 {
		lastEnd = ps.buffer[len(ps.buffer)-1].GetEnd()
	}

	if err := ps.ConsumeToken(&sparsing.Token{
		Kind:  sparsing.TokenKindEOF,
		Start: lastEnd,
		End:   lastEnd,
	}); err != nil {
		return err
	}

	ps.closed = true

	return nil
}

func (ps *parserStream[TToken, TNode]) Recover(offset int) sparsing.LexerStream {
	ps.err = nil
	ps.closed = false

	targetInStackIndex := -1

	for i, node := range ps.inStack {
		targetInStackIndex = i

		if offset < node.GetStartToken().GetStart().Offset {
			break
		}
	}

	targetInStackIndex--

	if targetInStackIndex < 0 {
		targetInStackIndex = 0
	}

	offending := ps.inStack[targetInStackIndex]
	ps.inStack = ps.inStack[:targetInStackIndex]

	offset = offending.GetStartToken().GetStart().Offset

	ps.LexerStream.Recover(offset)

	targetIndex := -1

	for i, tk := range ps.buffer {
		if tk.GetStart().Offset > offset {
			break
		}

		targetIndex = i
	}

	if targetIndex <= 0 {
		targetIndex = 0
	}

	ps.buffer = ps.buffer[:targetIndex]
	ps.bufferPosition = targetIndex

	return ps
}

func (ps *parserStream[TToken, TNode]) Reset() {
	ps.buffer = nil
	ps.bufferPosition = 0
	ps.err = nil
	ps.closed = false
}

func (ps *parserStream[TToken, TNode]) advance() (consumed int, err error) {
	if ps.err != nil {
		return 0, err
	}

	originalBufferPosition := ps.bufferPosition

	if originalBufferPosition >= len(ps.buffer) {
		return 0, nil
	}

	defer func() {
		var pe *sparsing.ParsingError

		if r := recover(); r != nil {
			wrapped := errors.Wrap(r, 1)

			if err != nil {
				err = multierror.Append(err, wrapped)
			} else {
				err = wrapped
			}
		}

		if err != nil && !errors.As(err, &pe) {
			pe = &sparsing.ParsingError{Err: err}
			pe.Position = ps.Position()

			if current := ps.CurrentNode(); any(current) != nil {
				pe.RecoverablePosition = current.GetStartToken().GetStart()
			} else if len(ps.buffer) > 0 {
				pe.RecoverablePosition = ps.buffer[originalBufferPosition].GetStart()
			}

			err = pe
		}

		consumed = ps.bufferPosition - originalBufferPosition

		if consumed < 0 {
			consumed = 0
		}

		if err != nil {
			ps.err = multierror.Append(ps.err, err)
		}
	}()

	if handler := ps.getTokenHandler(); handler != nil {
		if err = handler.ConsumeTokenStream(ps); err != nil {
			return
		}
	} else {
		err = sparsing.ErrNoTokenHandler
		return
	}

	return
}

func (ps *parserStream[TToken, TNode]) BeginComposite(s CompositeNode, start TToken) {
	s.SetStartToken(start)
	ps.PushInStack(s.(TNode))
}

func (ps *parserStream[TToken, TNode]) EndComposite(end TToken) {
	c := any(ps.CurrentNode()).(CompositeNode)

	c.SetEndToken(end)
	ps.PopToOutStack()
}

func (ps *parserStream[TToken, TNode]) PushTerminal(s TerminalNode, tk TToken) {
	if err := s.SetTerminalToken(tk); err != nil {
		panic(err)
	}

	ps.PushInStack(s.(TNode))
	ps.PopToOutStack()
}

func (ps *parserStream[TToken, TNode]) PushInStack(s TNode) {
	ps.inStack = append(ps.inStack, s)
}

func (ps *parserStream[TToken, TNode]) PopToOutStack() {
	if len(ps.inStack) == 0 {
		panic(errors.New("cannot pop empty input stack"))
	}

	current := ps.inStack[len(ps.inStack)-1]

	ps.outStack = append(ps.outStack, current)
	ps.inStack = ps.inStack[:len(ps.inStack)-1]

	if handler := ps.getNodeHandler(); handler != nil {
		if err := handler.ConsumeNode(ps, current); err != nil {
			panic(errors.New(err))
		}
	}
}

func (ps *parserStream[TToken, TNode]) ArgumentStackLength() int { return len(ps.outStack) }

func (ps *parserStream[TToken, TNode]) TryPopArgumentTo(ptr any) bool {
	if len(ps.outStack) == 0 {
		return false
	}

	v := reflect.ValueOf(ptr)
	v.Elem().Set(reflect.ValueOf(ps.PopArgument()))

	return true
}

func (ps *parserStream[TToken, TNode]) PushArgument(arg ...TNode) ParserStream[TToken, TNode] {
	ps.outStack = append(ps.outStack, arg...)

	return ps
}

func (ps *parserStream[TToken, TNode]) PopArgument() TNode {
	if len(ps.outStack) == 0 {
		panic(errors.New("cannot PopArgument empty outStack"))
	}

	n := ps.outStack[len(ps.outStack)-1]
	ps.outStack = ps.outStack[:len(ps.outStack)-1]

	return n
}

func (ps *parserStream[TToken, TNode]) ConsumeNextToken(kinds ...TokenKind) TToken {
	tk := ps.PeekToken(0)

	for _, kind := range kinds {
		if tk.GetKind() == kind {
			return ps.NextToken()
		}
	}

	panic(sparsing.ErrInvalidToken)
}

func (ps *parserStream[TToken, TNode]) NextToken() TToken {
	off := ps.bufferPosition

	if off >= len(ps.buffer) {
		panic(&sparsing.RollbackError{Err: sparsing.ErrEndOfStream})
	}

	token := ps.buffer[off]
	ps.bufferPosition++

	return token
}

func (ps *parserStream[TToken, TNode]) PeekToken(n int) Token {
	off := ps.bufferPosition + n

	if off >= len(ps.buffer) {
		return &sparsing.Token{
			Kind: sparsing.TokenKindEOS,
		}
	}

	return ps.buffer[off]
}

func (ps *parserStream[TToken, TNode]) PushTokenParser(handler ParserTokenHandler[TToken, TNode]) ParserStream[TToken, TNode] {
	ps.tokenHandlerStack = append(ps.tokenHandlerStack, handler)

	return ps
}

func (ps *parserStream[TToken, TNode]) PopTokenParser() ParserStream[TToken, TNode] {
	if len(ps.tokenHandlerStack) == 0 {
		panic("cannot pop token handler")
	}

	ps.tokenHandlerStack = ps.tokenHandlerStack[:len(ps.tokenHandlerStack)-1]

	return ps
}

func (ps *parserStream[TToken, TNode]) PushNodeConsumer(handler ParserNodeHandler[TToken, TNode]) ParserStream[TToken, TNode] {
	ps.nodeHandlerStack = append(ps.nodeHandlerStack, handler)

	return ps
}

func (ps *parserStream[TToken, TNode]) PopNodeConsumer() ParserStream[TToken, TNode] {
	if len(ps.nodeHandlerStack) == 0 {
		panic("cannot pop node handler")
	}

	ps.nodeHandlerStack = ps.nodeHandlerStack[:len(ps.nodeHandlerStack)-1]

	return ps
}

func (ps *parserStream[TToken, TNode]) getTokenHandler() ParserTokenHandler[TToken, TNode] {
	if len(ps.tokenHandlerStack) == 0 {
		return nil
	}

	return ps.tokenHandlerStack[len(ps.tokenHandlerStack)-1]
}

func (ps *parserStream[TToken, TNode]) getNodeHandler() ParserNodeHandler[TToken, TNode] {
	if len(ps.nodeHandlerStack) == 0 {
		return nil
	}

	return ps.nodeHandlerStack[len(ps.nodeHandlerStack)-1]
}
