package collectionsfx

import (
	obsfx2 "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type setExpressionHelper[T comparable] interface {
	SetListener[T]

	AddListener(listener obsfx2.InvalidationListener) setExpressionHelper[T]
	RemoveListener(listener obsfx2.InvalidationListener) setExpressionHelper[T]
	AddSetListener(listener SetListener[T]) setExpressionHelper[T]
	RemoveSetListener(listener SetListener[T]) setExpressionHelper[T]
}

type genericInvalidationSetExpressionHelper[T comparable] struct {
	invalidationListeners obsfx2.HasListenersBase[obsfx2.InvalidationListener]
	listListeners         obsfx2.HasListenersBase[SetListener[T]]
}

func newGenericInvalidationSetExpressionHelper[T comparable](invalidationListeners []obsfx2.InvalidationListener, listListeners []SetListener[T]) *genericInvalidationSetExpressionHelper[T] {
	g := &genericInvalidationSetExpressionHelper[T]{}

	for _, l := range invalidationListeners {
		g.AddListener(l)
	}

	for _, l := range listListeners {
		g.AddSetListener(l)
	}

	return g
}

func (g *genericInvalidationSetExpressionHelper[T]) OnSetChanged(ev SetChangeEvent[T]) {
	g.invalidationListeners.ForEachListener(func(l obsfx2.InvalidationListener) bool {
		l.OnInvalidated(ev.Set)

		return true
	})

	g.listListeners.ForEachListener(func(l SetListener[T]) bool {
		l.OnSetChanged(ev)

		return true
	})
}

func (g *genericInvalidationSetExpressionHelper[T]) AddListener(listener obsfx2.InvalidationListener) setExpressionHelper[T] {
	g.invalidationListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationSetExpressionHelper[T]) RemoveListener(listener obsfx2.InvalidationListener) setExpressionHelper[T] {
	g.invalidationListeners.RemoveListener(listener)

	return g
}

func (g *genericInvalidationSetExpressionHelper[T]) AddSetListener(listener SetListener[T]) setExpressionHelper[T] {
	g.listListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationSetExpressionHelper[T]) RemoveSetListener(listener SetListener[T]) setExpressionHelper[T] {
	g.listListeners.RemoveListener(listener)

	return g
}

type singleInvalidationSetExpressionHelper[T comparable] struct {
	listener obsfx2.InvalidationListener
}

func (s *singleInvalidationSetExpressionHelper[T]) AddSetListener(listener SetListener[T]) setExpressionHelper[T] {
	return newGenericInvalidationSetExpressionHelper[T]([]obsfx2.InvalidationListener{s.listener}, []SetListener[T]{listener})
}

func (s *singleInvalidationSetExpressionHelper[T]) RemoveSetListener(listener SetListener[T]) setExpressionHelper[T] {
	return s
}

func (s *singleInvalidationSetExpressionHelper[T]) AddListener(listener obsfx2.InvalidationListener) setExpressionHelper[T] {
	if s.listener == listener {
		return s
	}

	return newGenericInvalidationSetExpressionHelper[T]([]obsfx2.InvalidationListener{s.listener, listener}, nil)
}

func (s *singleInvalidationSetExpressionHelper[T]) RemoveListener(listener obsfx2.InvalidationListener) setExpressionHelper[T] {
	if s.listener == listener {
		return nil
	}

	return s
}

func (s *singleInvalidationSetExpressionHelper[T]) OnSetChanged(ev SetChangeEvent[T]) {
	s.listener.OnInvalidated(ev.Set)
}

type singleSetListenerExpressionHelper[T comparable] struct {
	listener SetListener[T]
}

func (s *singleSetListenerExpressionHelper[T]) OnSetChanged(ev SetChangeEvent[T]) {
	s.listener.OnSetChanged(ev)
}

func (s *singleSetListenerExpressionHelper[T]) AddListener(listener obsfx2.InvalidationListener) setExpressionHelper[T] {
	return newGenericInvalidationSetExpressionHelper[T]([]obsfx2.InvalidationListener{listener}, []SetListener[T]{s.listener})
}

func (s *singleSetListenerExpressionHelper[T]) RemoveListener(listener obsfx2.InvalidationListener) setExpressionHelper[T] {
	return s
}

func (s *singleSetListenerExpressionHelper[T]) AddSetListener(listener SetListener[T]) setExpressionHelper[T] {
	if s.listener == listener {
		return s
	}

	return newGenericInvalidationSetExpressionHelper[T](nil, []SetListener[T]{s.listener, listener})
}

func (s *singleSetListenerExpressionHelper[T]) RemoveSetListener(listener SetListener[T]) setExpressionHelper[T] {
	if s.listener == listener {
		return nil
	}

	return s
}
