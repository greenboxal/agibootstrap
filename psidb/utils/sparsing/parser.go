package sparsing

import (
	"reflect"

	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"
)

type ParserNodeHandler interface {
	ConsumeNode(ctx StreamingParserContext, node Node) error
}

type ParserNodeHandlerFunc func(ctx StreamingParserContext, node Node) error

func (p ParserNodeHandlerFunc) ConsumeNode(ctx StreamingParserContext, node Node) error {
	return p(ctx, node)
}

type ParserTokenHandler interface {
	ConsumeTokenStream(ctx StreamingParserContext) error
}

type ParserTokenHandlerFunc func(ctx StreamingParserContext) error

func (p ParserTokenHandlerFunc) ConsumeTokenStream(ctx StreamingParserContext) error {
	return p(ctx)
}

type StreamingParserContext interface {
	BeginComposite(s CompositeNode, start IToken)
	EndComposite(end IToken)
	PushTerminal(s TerminalNode, tk IToken)
	PushInStack(s Node)

	CurrentNode() Node
	CurrentPath() []Node
	ArgumentStackLength() int
	PopArgument() Node
	TryPopArgumentTo(ptr any) bool

	RemainingTokens() int
	ConsumeNextToken(kinds ...TokenKind) IToken
	NextToken() IToken
	PeekToken(n int) IToken

	PushTokenParser(handler ParserTokenHandler) ParserStream
	PopTokenParser() ParserStream

	PushNodeConsumer(handler ParserNodeHandler) ParserStream
	PopNodeConsumer() ParserStream
}

type ParserStream interface {
	LexerStream

	PushInStack(s Node)

	WriteTokens(tokens []IToken) (n int, err error)
	WriteToken(token IToken) error

	CurrentNode() Node
	ArgumentStackLength() int
	PopArgument() Node
	TryPopArgumentTo(ptr any) bool

	PushTokenParser(handler ParserTokenHandler) ParserStream
	PopTokenParser() ParserStream

	PushNodeConsumer(handler ParserNodeHandler) ParserStream
	PopNodeConsumer() ParserStream
}

type parserStream struct {
	LexerStream

	nodeHandlerStack  []ParserNodeHandler
	tokenHandlerStack []ParserTokenHandler

	buffer         []IToken
	bufferPosition int

	inStack  []Node
	outStack []Node

	err    error
	closed bool

	inParseLoop bool
}

func NewParserStream() ParserStream {
	ps := &parserStream{
		LexerStream: NewLexerStream(),
	}

	ps.LexerStream.PushLexerConsumer(ps)

	return ps
}

func (ps *parserStream) RemainingTokens() int { return len(ps.buffer) - ps.bufferPosition }

func (ps *parserStream) CurrentPath() []Node { return ps.inStack }
func (ps *parserStream) CurrentNode() Node {
	if len(ps.inStack) == 0 {
		return nil
	}

	return ps.inStack[len(ps.inStack)-1]
}

func (ps *parserStream) WriteTokens(tokens []IToken) (n int, err error) {
	if len(ps.tokenHandlerStack) == 0 {
		return 0, errors.New("no token handler set")
	}

	ps.buffer = append(ps.buffer, tokens...)

	return ps.advance()
}

func (ps *parserStream) ConsumeToken(token IToken) error {
	return ps.WriteToken(token)
}

func (ps *parserStream) WriteToken(token IToken) error {
	_, err := ps.WriteTokens([]IToken{token})

	if err != nil {
		return err
	}

	return nil
}

func (ps *parserStream) Close() error {
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

	if err := ps.ConsumeToken(&Token{
		Kind:  TokenKindEOF,
		Start: lastEnd,
		End:   lastEnd,
	}); err != nil {
		return err
	}

	ps.closed = true

	return nil
}

func (ps *parserStream) Recover(offset int) LexerStream {
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

func (ps *parserStream) Reset() {
	ps.buffer = nil
	ps.bufferPosition = 0
	ps.err = nil
	ps.closed = false
}

func (ps *parserStream) advance() (consumed int, err error) {
	if ps.err != nil {
		return 0, err
	}

	originalBufferPosition := ps.bufferPosition

	if originalBufferPosition >= len(ps.buffer) {
		return 0, nil
	}

	ps.inParseLoop = true

	defer func() {
		var pe *ParsingError

		ps.inParseLoop = false

		consumed = ps.bufferPosition - originalBufferPosition

		if consumed < 0 {
			consumed = 0
		}

		if r := recover(); r != nil {
			wrapped := errors.Wrap(r, 1)

			if err != nil {
				err = multierror.Append(err, wrapped)
			} else {
				err = wrapped
			}
		}

		if errors.Is(err, parserShiftError) {
			n, e := ps.advance()

			consumed += n
			err = e

			return
		}

		if err != nil && !errors.As(err, &pe) {
			pe = &ParsingError{Err: err}
			pe.Position = ps.Position()

			if current := ps.CurrentNode(); current != nil {
				pe.RecoverablePosition = current.GetStartToken().GetStart()
			} else if len(ps.buffer) > 0 {
				pe.RecoverablePosition = ps.buffer[originalBufferPosition].GetStart()
			}

			err = pe
		}

		if err != nil {
			ps.err = multierror.Append(ps.err, err)
		}
	}()

	for ps.RemainingTokens() > 0 {
		if handler := ps.getTokenHandler(); handler != nil {
			if err = handler.ConsumeTokenStream(ps); err != nil {
				return
			}

			if newHandler := ps.getTokenHandler(); newHandler != nil && newHandler != handler {
				continue
			}
		} else {
			err = ErrNoTokenHandler
			return
		}

		break
	}

	return
}

func (ps *parserStream) BeginComposite(s CompositeNode, start IToken) {
	if start == nil {
		start = ps.PeekToken(0)
	}

	s.SetStartToken(start)
	ps.PushInStack(s)
}

func (ps *parserStream) EndComposite(end IToken) {
	if end == nil {
		end = ps.buffer[ps.bufferPosition-1]
	}

	c := ps.CurrentNode().(CompositeNode)

	c.SetEndToken(end)
	ps.PopToOutStack()
}

func (ps *parserStream) PushTerminal(s TerminalNode, tk IToken) {
	if err := s.SetTerminalToken(tk); err != nil {
		panic(err)
	}

	ps.PushInStack(s)
	ps.PopToOutStack()
}

func (ps *parserStream) PushInStack(s Node) {
	shiftLexer, shiftParser := false, false

	ps.inStack = append(ps.inStack, s)

	if parseable, ok := s.(ParseableNode); ok {
		ps.PushTokenParser(parseable)
		shiftParser = true
	}

	if lexer, ok := s.(StreamingLexerHandler); ok {
		ps.PushLexerHandler(lexer)
		shiftLexer = true
	}

	if shiftLexer && ps.LexerStream.(*streamingLexerContext).inLexerLoop {
		panic(lexerShiftError)
	}

	if shiftParser && ps.inParseLoop {
		panic(parserShiftError)
	}
}

func (ps *parserStream) popFromInStack() Node {
	if len(ps.inStack) == 0 {
		panic(errors.New("cannot pop empty input stack"))
	}

	current := ps.inStack[len(ps.inStack)-1]
	ps.inStack = ps.inStack[:len(ps.inStack)-1]

	if _, ok := current.(ParseableNode); ok {
		ps.PopTokenParser()
	}

	if _, ok := current.(StreamingLexerHandler); ok {
		ps.PopLexerHandler()
	}

	return current
}

func (ps *parserStream) PopToOutStack() {
	current := ps.popFromInStack()

	ps.outStack = append(ps.outStack, current)

	if handler := ps.getNodeHandler(); handler != nil {
		if err := handler.ConsumeNode(ps, current); err != nil {
			panic(errors.New(err))
		}
	}
}

func (ps *parserStream) ArgumentStackLength() int { return len(ps.outStack) }

func (ps *parserStream) TryPopArgumentTo(ptr any) bool {
	if len(ps.outStack) == 0 {
		return false
	}

	v := reflect.ValueOf(ptr)
	v.Elem().Set(reflect.ValueOf(ps.PopArgument()))

	return true
}

func (ps *parserStream) PopArgument() Node {
	if len(ps.outStack) == 0 {
		panic(errors.New("cannot PopArgument empty outStack"))
	}

	n := ps.outStack[len(ps.outStack)-1]
	ps.outStack = ps.outStack[:len(ps.outStack)-1]

	return n
}

func (ps *parserStream) ConsumeNextToken(kinds ...TokenKind) IToken {
	tk := ps.PeekToken(0)

	for _, kind := range kinds {
		if tk.GetKind() == kind {
			return ps.NextToken()
		}
	}

	panic(ErrInvalidToken)
}

func (ps *parserStream) NextToken() IToken {
	off := ps.bufferPosition

	if off >= len(ps.buffer) {
		return &Token{
			Kind: TokenKindEOS,
		}
	}

	token := ps.buffer[off]
	ps.bufferPosition++

	return token
}

func (ps *parserStream) PeekToken(n int) IToken {
	off := ps.bufferPosition + n

	if off >= len(ps.buffer) {
		return &Token{
			Kind: TokenKindEOS,
		}
	}

	return ps.buffer[off]
}

func (ps *parserStream) PushTokenParser(handler ParserTokenHandler) ParserStream {
	ps.tokenHandlerStack = append(ps.tokenHandlerStack, handler)

	return ps
}

func (ps *parserStream) PopTokenParser() ParserStream {
	if len(ps.tokenHandlerStack) == 0 {
		panic("cannot pop token handler")
	}

	ps.tokenHandlerStack = ps.tokenHandlerStack[:len(ps.tokenHandlerStack)-1]

	return ps
}

func (ps *parserStream) PushNodeConsumer(handler ParserNodeHandler) ParserStream {
	ps.nodeHandlerStack = append(ps.nodeHandlerStack, handler)

	return ps
}

func (ps *parserStream) PopNodeConsumer() ParserStream {
	if len(ps.nodeHandlerStack) == 0 {
		panic("cannot pop node handler")
	}

	ps.nodeHandlerStack = ps.nodeHandlerStack[:len(ps.nodeHandlerStack)-1]

	return ps
}

func (ps *parserStream) getTokenHandler() ParserTokenHandler {
	if len(ps.tokenHandlerStack) == 0 {
		return nil
	}

	return ps.tokenHandlerStack[len(ps.tokenHandlerStack)-1]
}

func (ps *parserStream) getNodeHandler() ParserNodeHandler {
	if len(ps.nodeHandlerStack) == 0 {
		return nil
	}

	return ps.nodeHandlerStack[len(ps.nodeHandlerStack)-1]
}

func ConsumeAs[T Node](p ParserStream) T {
	var result T

	if v, ok := p.CurrentNode().(T); ok {
		return v
	}

	if !p.TryPopArgumentTo(&result) {
		panic("cannot pop argument")
	}

	return result
}
