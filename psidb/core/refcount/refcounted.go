package refcount

import "sync"

type IRefCounted interface {
	GetRefCountObjectSlot() IObjectSlot
}

type RefCounted[T any] struct {
	IRefCounted

	ObjectSlot[T]
}

func (r RefCounted[T]) GetRefCountObjectSlot() IObjectSlot { return r.ObjectSlot }

func MakeRefCounted[T any](rc *RefCounted[T], value T, finalizer ReferenceFinalizer[T]) {
	rc.ObjectSlot = &objectSlot[T]{
		value:     value,
		finalizer: finalizer,
		refs:      -1,
	}
}

type IObjectSlot interface {
	IsValid() bool
}

type ObjectSlot[T any] interface {
	IObjectSlot

	Value() T
	Ref() ObjectHandle[T]
	Unref(oh ObjectHandle[T])
	SetFinalizer(oh ReferenceFinalizer[T])
}

type ObjectHandle[T any] interface {
	IsValid() bool
	Get() T
	Release()
}

type ReferenceFinalizer[T any] func(obj ObjectSlot[T], value T, next ReferenceFinalizer[T])

type objectSlot[T any] struct {
	m         sync.Mutex
	refs      int
	value     T
	finalizer ReferenceFinalizer[T]
}

func (s *objectSlot[T]) IsValid() bool {
	return s != nil && (s.refs > 0 || s.refs == -1)
}

func (s *objectSlot[T]) Value() T {
	return s.value
}

func (s *objectSlot[T]) Ref() ObjectHandle[T] {
	s.m.Lock()
	defer s.m.Unlock()

	if !s.IsValid() {
		panic("objectSlot.Ref: !IsValid()")
	}

	if s.refs == -1 {
		s.refs = 1
	} else {
		s.refs++
	}

	return &objectHandle[T]{
		obj: s,
	}
}

func (s *objectSlot[T]) Unref(oh ObjectHandle[T]) {
	s.m.Lock()
	if s.refs == 0 {
		s.m.Unlock()
		panic("objectSlot.Unref: refs == 0")
	}

	if s.refs > 0 {
		s.refs--
	}
	s.m.Unlock()

	if s.refs == 0 {
		s.Close()
	}
}

func (s *objectSlot[T]) SetFinalizer(f ReferenceFinalizer[T]) {
	s.m.Lock()
	defer s.m.Unlock()

	previous := s.finalizer

	s.finalizer = func(obj ObjectSlot[T], value T, _ ReferenceFinalizer[T]) {
		f(obj, value, previous)
	}
}

func (s *objectSlot[T]) Close() {
	var empty T

	s.m.Lock()
	defer s.m.Unlock()

	if s.refs != 0 {
		panic("objectSlot.Close: refs != 0")
	}

	if s.finalizer != nil {
		s.finalizer(s, s.value, nil)
	}

	s.value = empty
}

type objectHandle[T any] struct {
	m   sync.Mutex
	obj *objectSlot[T]
}

func (oh *objectHandle[T]) IsValid() bool {
	return oh != nil && oh.obj != nil
}

func (oh *objectHandle[T]) Get() (_ T) {
	if !oh.IsValid() {
		return
	}

	return oh.obj.value
}

func (oh *objectHandle[T]) Release() {
	if !oh.IsValid() {
		return
	}

	oh.m.Lock()
	defer oh.m.Unlock()

	oh.obj.Unref(oh)
	oh.obj = nil
}
