package sparsing

import (
	"fmt"
	"io"
	"regexp"

	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"
)

const RuneEOF = rune(0)
const RuneEOS = rune(1)

type LexerStream interface {
	io.Writer
	io.Closer

	Position() Position
	Recover(offset int) LexerStream

	PushLexerConsumer(handler LexerTokenHandler) LexerStream
	PopLexerConsumer() LexerStream

	PushLexerHandler(handler StreamingLexerHandler) LexerStream
	PopLexerHandler() LexerStream

	Clone() LexerStream
}

type LexerTokenHandler interface {
	ConsumeToken(token IToken) error
}

type LexerTokenHandlerFunc func(token IToken) error

func (f LexerTokenHandlerFunc) ConsumeToken(token IToken) error {
	return f(token)
}

type LexerMatchOperations interface {
	NextChar() rune
	LA(i int) rune

	PeekMatch(pattern string) bool
	TryMatch(pattern string) bool
	Match(pattern string)

	PeekMatchRegexp(pattern *regexp.Regexp) ([]string, bool)
	TryMatchRegexp(pattern *regexp.Regexp) ([]string, bool)
	MatchRegexp(pattern *regexp.Regexp) []string

	Remaining() int
	PeekBuffer() string
}

type LexerAssemblyOperations interface {
	PushEnd()
	AppendValue(s string)
	PushSingle(kind TokenKind, s string)
	PushStart(kindString TokenKind, s string)

	CurrentToken() *Token
}

type StreamingLexerHandler interface {
	ConsumeStream(ctx StreamingLexerContext) error
}

type StreamingLexerHandlerFunc func(ctx StreamingLexerContext) error

func (f StreamingLexerHandlerFunc) ConsumeStream(ctx StreamingLexerContext) error {
	return f(ctx)
}

type StreamingLexerContext interface {
	io.Reader
	io.ReaderAt
	io.Seeker

	LexerMatchOperations
	LexerAssemblyOperations

	PushLexerConsumer(handler LexerTokenHandler) LexerStream
	PopLexerConsumer() LexerStream

	PushLexerHandler(handler StreamingLexerHandler) LexerStream
	PopLexerHandler() LexerStream
}

type streamingLexerContext struct {
	lexerHandlerStack []StreamingLexerHandler
	tokenHandlerStack []LexerTokenHandler

	buffer         []rune
	bufferPosition int

	pos Position
	err error

	nextToken Token
	nextIndex int

	closed bool

	inLexerLoop bool
}

func (l *streamingLexerContext) PeekMatch(pattern string) bool {
	if l.Remaining() < len(pattern) {
		return false
	}

	return l.PeekBuffer()[:len(pattern)] == pattern
}

func (l *streamingLexerContext) TryMatch(pattern string) bool {
	if !l.PeekMatch(pattern) {
		return false
	}

	l.bufferPosition += len(pattern)

	return true
}

func (l *streamingLexerContext) Match(pattern string) {
	if !l.TryMatch(pattern) {
		panic(l.PushError(fmt.Errorf("expected %s", pattern)))
	}
}

func (l *streamingLexerContext) PeekMatchRegexp(pattern *regexp.Regexp) ([]string, bool) {
	prefix, complete := pattern.LiteralPrefix()

	if !l.TryMatch(prefix) {
		return nil, false
	}

	if complete {
		return nil, true
	}

	match := pattern.FindStringSubmatch(l.PeekBuffer())

	if match == nil {
		return nil, false
	}

	return match, true
}

func (l *streamingLexerContext) TryMatchRegexp(pattern *regexp.Regexp) ([]string, bool) {
	match, ok := l.PeekMatchRegexp(pattern)

	if !ok {
		return nil, false
	}

	l.bufferPosition += len(match[0])

	return match, true
}

func (l *streamingLexerContext) MatchRegexp(pattern *regexp.Regexp) []string {
	match, ok := l.TryMatchRegexp(pattern)

	if !ok {
		panic(l.PushError(fmt.Errorf("expected %s", pattern)))
	}

	return match
}

func NewLexerStream() LexerStream {
	return &streamingLexerContext{}
}

func NewLexerStreamWithHandlers(lexerHandler StreamingLexerHandler, tokenHandler LexerTokenHandler) LexerStream {
	sc := &streamingLexerContext{}

	if lexerHandler != nil {
		sc.lexerHandlerStack = append(sc.lexerHandlerStack, lexerHandler)
	}

	if tokenHandler != nil {
		sc.tokenHandlerStack = append(sc.tokenHandlerStack, tokenHandler)
	}

	return sc
}

func NewLexerStreamWithHandlerStack(lexerHandler StreamingLexerHandler, tokenHandler []LexerTokenHandler) LexerStream {
	return &streamingLexerContext{
		lexerHandlerStack: []StreamingLexerHandler{lexerHandler},
		tokenHandlerStack: tokenHandler,
	}
}

func (l *streamingLexerContext) Position() Position   { return l.pos }
func (l *streamingLexerContext) CurrentToken() *Token { return &l.nextToken }

func (l *streamingLexerContext) PeekBuffer() string               { return string(l.buffer[l.bufferPosition:]) }
func (l *streamingLexerContext) Remaining() int                   { return len(l.buffer) - l.bufferPosition }
func (l *streamingLexerContext) Read(p []byte) (n int, err error) { return l.ReadAt(p, 0) }
func (l *streamingLexerContext) ReadAt(p []byte, off int64) (n int, err error) {
	offset := l.bufferPosition + int(off)

	if offset >= len(l.buffer) {
		return 0, io.EOF
	}

	count := copy(p, string(l.buffer[offset:]))

	l.bufferPosition += count

	return count, nil
}

func (l *streamingLexerContext) LA(n int) rune {
	if l.closed {
		return RuneEOF
	}

	index := l.bufferPosition + n

	if index >= len(l.buffer) {
		return RuneEOS
	}

	return l.buffer[index]
}

func (l *streamingLexerContext) NextChar() rune {
	r := l.getNextChar()

	if r == RuneEOS {
		panic(&RollbackError{Err: ErrEndOfStream})
	} else if r == '\n' {
		l.pos.Line++
		l.pos.Column = 0
	} else {
		l.pos.Column++
	}

	return r
}
func (l *streamingLexerContext) getNextChar() rune {
	if l.closed {
		return RuneEOF
	}

	index := l.bufferPosition

	if index >= len(l.buffer) {
		return RuneEOS
	}

	l.pos.Offset = l.bufferPosition
	l.bufferPosition++

	r := l.buffer[index]

	return r
}

func (l *streamingLexerContext) Recover(offset int) LexerStream {
	if offset < 0 || offset > len(l.buffer) {
		panic(errors.New("invalid offset"))
	}

	l.err = nil
	l.closed = false

	if l.nextToken.Kind != TokenKindInvalid {
		if l.nextToken.Start.Offset >= offset {
			l.nextToken = Token{}
		} else {
			removed := offset - l.nextToken.Start.Offset

			if removed >= len(l.nextToken.Value) {
				l.nextToken = Token{}
			} else {
				l.nextToken.Value = l.nextToken.Value[removed:]
			}
		}
	}

	if _, err := l.Seek(int64(offset), io.SeekStart); err != nil {
		panic(err)
	}

	l.buffer = l.buffer[:l.bufferPosition]

	return l
}

func (l *streamingLexerContext) Seek(offset int64, whence int) (int64, error) {
	if l.closed {
		return 0, io.ErrClosedPipe
	}

	if l.err != nil {
		return 0, l.err
	}

	var err error

	switch whence {
	case io.SeekStart:
		l.bufferPosition = int(offset)
	case io.SeekCurrent:
		l.bufferPosition += int(offset)
	case io.SeekEnd:
		l.bufferPosition = len(l.buffer) + int(offset)
	default:
		err = errors.New("invalid whence")
	}

	if l.bufferPosition < 0 {
		l.bufferPosition = 0
	}

	if l.bufferPosition > len(l.buffer) {
		l.bufferPosition = len(l.buffer)
	}

	l.pos.Offset = l.bufferPosition

	return int64(l.bufferPosition), err
}

func (l *streamingLexerContext) Write(p []byte) (n int, err error) {
	if l.closed {
		l.closed = false
	}

	if l.err != nil {
		return 0, l.err
	}

	appendOffset := len(l.buffer)
	l.buffer = append(l.buffer, []rune(string(p))...)

	n, err = l.advance(nil)

	if n < len(p) {
		removeLen := len(p) - n
		updated := l.buffer[:appendOffset]
		updated = append(updated, l.buffer[appendOffset+removeLen:]...)
		l.buffer = updated
	}

	return n, err
}

func (l *streamingLexerContext) Reset() {
	l.buffer = []rune{}
	l.bufferPosition = 0
	l.pos = Position{}
	l.err = nil
	l.nextToken = Token{}
	l.nextIndex = 0
	l.closed = false
}

func (l *streamingLexerContext) Close() error {
	if l.closed {
		return l.err
	}

	if l.err != nil {
		return l.err
	}

	_, err := l.Write([]byte(string(RuneEOF)))

	if err != nil {
		return err
	}

	l.closed = true

	return nil
}

var lexerShiftError = errors.New("lexer shift error")
var parserShiftError = errors.New("parser shift error")

func (l *streamingLexerContext) advance(postHook func()) (consumed int, err error) {
	if l.err != nil {
		return 0, err
	}

	recoverablePosition := l.pos
	originalBufferPosition := l.bufferPosition

	if originalBufferPosition >= len(l.buffer) {
		return 0, nil
	}

	l.inLexerLoop = true

	defer func() {
		var parsingError *ParsingError
		var rollbackError *RollbackError

		l.inLexerLoop = false

		if r := recover(); r != nil {
			wrapped := errors.Wrap(r, 1)

			if err != nil {
				err = multierror.Append(err, wrapped)
			} else {
				err = wrapped
			}
		}

		if errors.Is(err, lexerShiftError) {
			n, e := l.advance(postHook)

			consumed += n
			err = e

			return
		}

		if errors.As(err, &rollbackError) {
			if errors.Is(rollbackError.Err, ErrEndOfStream) {
				err = nil
			} else {
				err = rollbackError.Err
			}
		}

		if err != nil && !errors.As(err, &parsingError) {
			parsingError = &ParsingError{
				Err:                 err,
				Position:            l.Position(),
				RecoverablePosition: recoverablePosition,
			}

			err = parsingError
		}

		consumed = l.bufferPosition - originalBufferPosition

		if consumed < 0 {
			consumed = 0
		}

		if err != nil {
			l.err = multierror.Append(l.err, err)
		}
	}()

	if handler := l.getLexerHandler(); handler != nil {
		if err = handler.ConsumeStream(l); err != nil {
			return
		}
	} else {
		panic(l.PushError(ErrNoLexerHandler))
	}

	if postHook != nil {
		postHook()
	}

	return
}

func (l *streamingLexerContext) PushError(err error) error {
	var pe *ParsingError

	if !errors.As(err, &pe) {
		pe = &ParsingError{
			Err:      errors.Wrap(err, 1),
			Position: l.pos,
		}
	}

	panic(errors.New(pe))

	return err
}

func (l *streamingLexerContext) AppendValue(s string) {
	l.nextToken.Value += s
}

func (l *streamingLexerContext) PushStart(kind TokenKind, s string) {
	l.nextToken.Kind = kind
	l.nextToken.Start = l.pos
	l.nextToken.End = l.pos
	l.nextToken.Value = s
}

func (l *streamingLexerContext) PushEnd() {
	if l.nextToken.Kind == TokenKindInvalid {
		return
	}

	l.nextToken.End = l.pos
	if l.nextToken.Start == l.pos {
		l.nextToken.End.Offset += len(l.nextToken.Value)
		l.nextToken.End.Column += len(l.nextToken.Value)
	}
	l.nextToken.Index = l.nextIndex

	l.nextIndex++
	cloned := l.nextToken

	l.nextToken.Kind = TokenKindInvalid
	l.nextToken.Start = l.nextToken.End
	l.nextToken.Value = ""

	if handler := l.getTokenHandler(); handler != nil {

		if err := handler.ConsumeToken(&cloned); err != nil {
			panic(l.PushError(err))
		}
	} else {
		panic(l.PushError(ErrNoTokenHandler))
	}
}

func (l *streamingLexerContext) PushSingle(kind TokenKind, s string) {
	l.nextToken = Token{
		Kind:  kind,
		Value: s,
		Start: l.pos,
		End: Position{
			Offset: l.pos.Offset + len(s),
			Line:   l.pos.Line,
			Column: l.pos.Column + len(s),
		},
	}

	l.PushEnd()
}

func (l *streamingLexerContext) PushLexerConsumer(handler LexerTokenHandler) LexerStream {
	l.tokenHandlerStack = append(l.tokenHandlerStack, handler)

	return l
}

func (l *streamingLexerContext) PopLexerConsumer() LexerStream {
	if len(l.tokenHandlerStack) == 0 {
		panic(errors.New("cannot pop empty tokenHandlerStack"))
	}

	l.tokenHandlerStack = l.tokenHandlerStack[:len(l.tokenHandlerStack)-1]

	return l
}

func (l *streamingLexerContext) PushLexerHandler(handler StreamingLexerHandler) LexerStream {
	l.lexerHandlerStack = append(l.lexerHandlerStack, handler)

	return l
}

func (l *streamingLexerContext) PopLexerHandler() LexerStream {
	if len(l.lexerHandlerStack) == 0 {
		panic(errors.New("cannot pop empty lexerHandlerStack"))
	}

	l.lexerHandlerStack = l.lexerHandlerStack[:len(l.lexerHandlerStack)-1]

	return l
}

func (l *streamingLexerContext) Clone() LexerStream {
	return &streamingLexerContext{
		lexerHandlerStack: l.lexerHandlerStack,
		tokenHandlerStack: l.tokenHandlerStack,
		buffer:            l.buffer,
		bufferPosition:    l.bufferPosition,
		pos:               l.pos,
		err:               l.err,
		nextToken:         l.nextToken,
		nextIndex:         l.nextIndex,
		closed:            l.closed,
	}
}

func (l *streamingLexerContext) getLexerHandler() StreamingLexerHandler {
	if len(l.lexerHandlerStack) == 0 {
		return nil
	}

	return l.lexerHandlerStack[len(l.lexerHandlerStack)-1]
}

func (l *streamingLexerContext) getTokenHandler() LexerTokenHandler {
	if len(l.tokenHandlerStack) == 0 {
		return nil
	}

	return l.tokenHandlerStack[len(l.tokenHandlerStack)-1]
}
