package stdlib

import (
	"context"
	"reflect"

	"github.com/go-errors/errors"
	"golang.org/x/exp/maps"
)

var ErrStopIteration = errors.New("stop iteration")

type Result[T any] struct {
	Value T
	Err   error
}

type KeyValue[K comparable, V any] struct {
	K K
	V V
}

type Iterator[T any] interface {
	Next() bool
	Value() T
}

type Iterable[T any] interface {
	Iterator() Iterator[T]
}

type Reducer[T any, U any] interface {
	Reduce(Iterator[T]) U
}

func Concat[T any](iterators ...Iterator[T]) Iterator[T] {
	return &concatIterator[T]{iterators: iterators}
}

func Filter[IT Iterator[T], T any](iter IT, fn func(T) bool) Iterator[T] {
	return &filterIterator[T]{iter: iter, fn: fn}
}

func FilterIsInstance[T, U any](iter Iterator[T]) Iterator[U] {
	return &mapFilterIterator[T, U]{iter: iter, typ: reflect.TypeOf((*U)(nil)).Elem()}
}

func Map[IT Iterator[T], T, U any](iter IT, fn func(T) U) Iterator[U] {
	return &mapIterator[T, U]{iter: iter, fn: fn}
}

func FlatMap[IT Iterator[T], T, U any](iter IT, fn func(T) Iterator[U]) Iterator[U] {
	return &flatMapIterator[T, U]{src: iter, fn: fn}
}

func Count[IT Iterator[T], T any](iter IT) int {
	var count int

	for iter.Next() {
		count++
	}

	return count
}

func ToMap[IT Iterator[KeyValue[K, V]], K comparable, V any](iter IT) map[K]V {
	result := make(map[K]V)

	for iter.Next() {
		pair := iter.Value()
		result[pair.K] = pair.V
	}

	return result
}

func ToSlice[T any](iter Iterator[T]) []T {
	var result []T

	for iter.Next() {
		result = append(result, iter.Value())
	}

	return result
}

func AsChannel[IT Iterator[T], T any](iter IT) <-chan T {
	ch := make(chan T)

	go func() {
		defer close(ch)

		for iter.Next() {
			ch <- iter.Value()
		}
	}()

	return ch
}

func AsBufferedChannel[IT Iterator[T], T any](iter IT, size int) <-chan T {
	ch := make(chan T, size)

	go func() {
		defer close(ch)

		for iter.Next() {
			ch <- iter.Value()
		}
	}()

	return ch
}

func PumpToChannel[IT Iterator[T], T any](ctx context.Context, iter IT, ch chan<- T) {
	for iter.Next() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ch <- iter.Value()

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func ToSliceWithCapacity[IT Iterator[T], T any](iter IT, capacity int) []T {
	result := make([]T, 0, capacity)

	for iter.Next() {
		result = append(result, iter.Value())
	}

	return result
}

type funcIterator[T any] struct {
	fn      func() (T, bool)
	current T
}

func (f *funcIterator[T]) Next() bool {
	value, ok := f.fn()

	if !ok {
		return false
	}

	f.current = value

	return true
}

func (f *funcIterator[T]) Value() T {
	return f.current
}

func NewIterator[T any](fn func() (T, bool)) Iterator[T] {
	return &funcIterator[T]{fn: fn}
}

func FromChannel[T any](ch <-chan T) Iterator[T] {
	return &channelIterator[T]{ch: ch}
}

func FromSlice[T any](slice []T) Iterator[T] {
	return &sliceIterator[T]{slice: slice}
}

func FromMap[K comparable, V any](m map[K]V) Iterator[KeyValue[K, V]] {
	return &mapEntryIterator[K, V]{m: m, keys: maps.Keys(m)}
}

func Buffer[T any](iter Iterator[T], size int) Iterator[Iterator[T]] {
	return &chunkIterator[T]{iter: iter, size: size}
}

func Buffered[T any](iter Iterator[T], size int) Iterator[T] {
	return &buffered[T]{iter: iter, buffer: make([]T, size)}
}

func OnEach[T any](iter Iterator[T], fn func(T)) Iterator[T] {
	return &onEachIterator[T]{iter: iter, fn: fn}
}

func Fold[IT Iterator[T], T any, U any](iter IT, initial U, fn func(U, T) U) U {
	for iter.Next() {
		initial = fn(initial, iter.Value())
	}

	return initial
}

func Reduce[IT Iterator[T], T any, U any](iter IT, reducer func(Iterator[T]) U) U {
	return reducer(iter)
}

func Scan[IT Iterator[T], T any, U any](iter IT, initial U, fn func(U, T) U) Iterator[U] {
	return &scanIterator[T, U]{iter: iter, current: initial, fn: fn}
}

func ForEach[IT Iterator[T], T any](iter IT, fn func(T)) {
	for iter.Next() {
		fn(iter.Value())
	}
}

func ForEachContext[IT Iterator[T], T any](ctx context.Context, iter IT, fn func(context.Context, T)) {
	for iter.Next() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fn(ctx, iter.Value())

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func Cancellable[T any](ctx context.Context, iter Iterator[T]) Iterator[T] {
	return &cancellableIterator[T]{ctx: ctx, iter: iter}
}

func Safe[IT Iterator[T], T any](iter IT) Iterator[Result[T]] {
	return &safeIterator[T]{iter: iter}
}

type mapFilterIterator[T, U any] struct {
	iter    Iterator[T]
	typ     reflect.Type
	current U
}

func (m *mapFilterIterator[T, U]) Next() bool {
	for m.iter.Next() {
		v := reflect.ValueOf(m.iter.Value())

		if v.Type().ConvertibleTo(m.typ) {
			m.current = v.Convert(m.typ).Interface().(U)
			return true
		}
	}

	return false
}

func (m *mapFilterIterator[T, U]) Value() U {
	return m.current
}

type safeIterator[T any] struct {
	iter    Iterator[T]
	current Result[T]
	empty   T
}

func (s *safeIterator[T]) Next() bool {
	defer func() {
		if err := recover(); err != nil {
			if err == ErrStopIteration {
				s.current.Value = s.empty
				s.current.Err = nil
				return
			}

			if err, ok := err.(error); ok {
				s.current.Err = err
			} else {
				s.current.Err = errors.Wrap(err, 1)
			}
		}
	}()

	s.current.Value = s.empty

	if !s.iter.Next() {
		return false
	}

	s.current.Value = s.iter.Value()

	return true
}

func (s *safeIterator[T]) Value() Result[T] {
	return s.current
}

type cancellableIterator[T any] struct {
	ctx     context.Context
	iter    Iterator[T]
	current T
}

func (c *cancellableIterator[T]) Next() bool {
	defer func() {
		if err := recover(); err != nil {
			if err == ErrStopIteration {
				return
			}

			panic(err)
		}
	}()

	if !c.iter.Next() {
		return false
	}

	select {
	case <-c.ctx.Done():
		return false
	default:
	}

	c.current = c.iter.Value()

	return true
}

func (c *cancellableIterator[T]) Value() T {
	return c.current
}

type onEachIterator[T any] struct {
	iter Iterator[T]
	fn   func(T)
}

func (o *onEachIterator[T]) Next() bool {
	if !o.iter.Next() {
		return false
	}

	o.fn(o.iter.Value())

	return true
}

func (o *onEachIterator[T]) Value() T {
	return o.iter.Value()
}

type mapEntryIterator[K comparable, V any] struct {
	m       map[K]V
	keys    []K
	current KeyValue[K, V]
}

func (m *mapEntryIterator[K, V]) Next() bool {
	if len(m.keys) == 0 {
		return false
	}

	key := m.keys[0]
	m.keys = m.keys[1:]
	m.current = KeyValue[K, V]{K: key, V: m.m[key]}

	return true
}

func (m *mapEntryIterator[K, V]) Value() KeyValue[K, V] {
	return m.current
}

type scanIterator[T, U any] struct {
	iter    Iterator[T]
	current U
	fn      func(U, T) U
}

func (s *scanIterator[T, U]) Next() bool {
	if !s.iter.Next() {
		return false
	}

	s.current = s.fn(s.current, s.iter.Value())

	return true
}

func (s *scanIterator[T, U]) Value() U {
	return s.current
}

type channelIterator[T any] struct {
	ch      <-chan T
	current T
}

func (c *channelIterator[T]) Next() bool {
	var ok bool

	c.current, ok = <-c.ch

	return ok
}

func (c *channelIterator[T]) Value() T {
	return c.current
}

type buffered[T any] struct {
	iter   Iterator[T]
	buffer []T
	first  int
	last   int
}

func (b *buffered[T]) Next() bool {
	missing := len(b.buffer) - (b.last - b.first)

	for i := 0; i < missing; i++ {
		if !b.iter.Next() {
			return false
		}

		b.buffer[b.last] = b.iter.Value()
		b.last++

		if b.last >= len(b.buffer) {
			b.last = 0
		}
	}

	if b.first == b.last {
		return false
	}

	b.first++

	if b.first >= len(b.buffer) {
		b.first = 0
	}

	return true
}

func (b *buffered[T]) Value() (def T) {
	if b.first == b.last {
		return
	}

	return b.buffer[b.first]
}

type mapIterator[T, U any] struct {
	iter  Iterator[T]
	fn    func(T) U
	value U
}

func (m *mapIterator[T, U]) Next() bool {
	if !m.iter.Next() {
		return false
	}

	m.value = m.fn(m.iter.Value())

	return true
}

func (m *mapIterator[T, U]) Value() U {
	return m.value
}

type flatMapIterator[T, U any] struct {
	src     Iterator[T]
	fn      func(T) Iterator[U]
	current Iterator[U]
	value   U
}

func (f *flatMapIterator[T, U]) Next() bool {
	for {
		if f.current == nil {
			if !f.src.Next() {
				return false
			}

			f.current = f.fn(f.src.Value())
		}

		if !f.current.Next() {
			f.current = nil

			continue
		}

		f.value = f.current.Value()

		return true
	}
}

type concatIterator[T any] struct {
	iterators []Iterator[T]
	current   Iterator[T]
}

func (c *concatIterator[T]) Next() bool {
	for {
		if c.current == nil {
			if len(c.iterators) == 0 {
				return false
			}

			c.current = c.iterators[0]
			c.iterators = c.iterators[1:]
		}

		if !c.current.Next() {
			c.current = nil

			continue
		}

		return true
	}
}

func (c *concatIterator[T]) Value() (def T) {
	if c.current == nil {
		return
	}

	return c.current.Value()
}

func (f *flatMapIterator[T, U]) Value() U {
	return f.value
}

type filterIterator[T any] struct {
	iter Iterator[T]
	fn   func(T) bool
}

func (f *filterIterator[T]) Next() bool {
	for {
		if !f.iter.Next() {
			return false
		}

		if !f.fn(f.iter.Value()) {
			continue
		}

		return true
	}
}

func (f *filterIterator[T]) Value() T {
	return f.iter.Value()
}

type chunkIterator[T any] struct {
	iter  Iterator[T]
	size  int
	chunk Iterator[T]
}

type sliceIterator[T any] struct {
	slice   []T
	current T
}

func (s *sliceIterator[T]) Next() bool {
	if len(s.slice) == 0 {
		return false
	}

	s.current = s.slice[0]
	s.slice = s.slice[1:]

	return true
}

func (s *sliceIterator[T]) Value() (def T) {
	return s.current
}

func (c *chunkIterator[T]) Next() bool {
	var chunk []T

	for c.iter.Next() {
		chunk = append(chunk, c.iter.Value())

		if len(chunk) == c.size {
			break
		}
	}

	if len(chunk) == 0 {
		return false
	}

	c.chunk = &sliceIterator[T]{slice: chunk}

	return true
}

func (c *chunkIterator[T]) Value() Iterator[T] {
	return c.chunk
}
