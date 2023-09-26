package turing

import "golang.org/x/exp/slices"

type OperationStack = MutableStack[Value]
type ValueStack = MutableStack[Value]

type Stack[T any] interface {
	InsertAt(idx int, v T)
	Deque() T
	Append(v ...T)
	Push(v T)
	Pop() T
	Peek() T

	LA(i int) T
	Len() int
	Clone() Stack[T]
	ToSlice() []T
}

type MutableStack[T any] struct {
	items []T
}

func (s *MutableStack[T]) InsertAt(idx int, v T) {
	s.items = slices.Insert(s.items, idx, v)
}

func (s *MutableStack[T]) Deque() T {
	if len(s.items) == 0 {
		panic("stack is empty")
	}

	v := s.items[0]
	s.items = s.items[1:]

	return v
}

func (s *MutableStack[T]) Append(v ...T) {
	s.items = append(s.items, v...)
}

func (s *MutableStack[T]) Push(v T) {
	s.items = append(s.items, v)
}

func (s *MutableStack[T]) Pop() T {
	if len(s.items) == 0 {
		panic("stack is empty")
	}

	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]

	return v
}

func (s *MutableStack[T]) Peek() (empty T) {
	v, _ := s.LA(0)
	return v
}

func (s *MutableStack[T]) LA(i int) (empty T, _ bool) {
	index := len(s.items) - 1 - i

	if index < 0 {
		return empty, false
	}

	return s.items[index], true
}

func (s *MutableStack[T]) Len() int {
	return len(s.items)
}

func (s *MutableStack[T]) Clone() *MutableStack[T] {
	return &MutableStack[T]{
		items: append([]T{}, s.items...),
	}
}

func (s *MutableStack[T]) ToSlice() []T {
	return append([]T{}, s.items...)
}

func (s *MutableStack[T]) All() []T {
	return s.items
}
