package lex

import (
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type BasicTokenBuffer interface {
	Len() int
}

type TokenConverter[T, U any] func(T) (U, error)

type TokenBuffer[T any] interface {
	BasicTokenBuffer

	iterators.Iterable[T]

	Position() int
	SetPosition(pos int)

	Write(data []T) error
	WriteValue(value T) error

	Copy(dst TokenBuffer[T]) error
	Slice(start, end int) TokenBuffer[T]

	ToSlice(dst []T) []T

	Reset()
}

type CharTokenBuffer = TokenBufferBase[rune]
type ByteTokenBuffer = TokenBufferBase[byte]
type IntTokenBuffer = TokenBufferBase[int]

type StringTokenBuffer struct {
	TokenBufferBase[byte]
}

func (s *StringTokenBuffer) String() string {
	return string(s.buffer)
}

func (s *StringTokenBuffer) StringSlice(start, end int) string {
	return string(s.buffer[start:end])
}

type TokenBufferBase[T any] struct {
	buffer []T
	pos    int
}

func (t *TokenBufferBase[T]) Len() int                        { return len(t.buffer) }
func (t *TokenBufferBase[T]) Iterator() iterators.Iterator[T] { return iterators.FromSlice(t.buffer) }

func (t *TokenBufferBase[T]) Position() int { return t.pos }

func (t *TokenBufferBase[T]) SetPosition(pos int) {
	if pos < 0 || pos > len(t.buffer) {
		panic("invalid position")
	}

	t.pos = pos
}

func (t *TokenBufferBase[T]) Write(data []T) error {
	t.buffer = slices.Insert(t.buffer, t.pos, data...)

	return nil
}

func (t *TokenBufferBase[T]) WriteValue(value T) error {
	return t.Write([]T{value})
}

func (t *TokenBufferBase[T]) Copy(dst TokenBuffer[T]) error {
	return dst.Write(t.ToSlice(nil))
}

func (t *TokenBufferBase[T]) Slice(start, end int) TokenBuffer[T] {
	return &TokenBufferBase[T]{
		buffer: t.buffer[start:end],
		pos:    0,
	}
}

func (t *TokenBufferBase[T]) ToSlice(dst []T) []T {
	dst = append(dst, t.buffer...)

	return dst
}

func (t *TokenBufferBase[T]) Reset() {
	t.buffer = t.buffer[0:0]
	t.pos = 0
}
