package collectionsfx

import (
	"reflect"

	obsfx2 "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type listExpressionHelper[T any] interface {
	ListListener[T]

	AddListener(listener obsfx2.InvalidationListener) listExpressionHelper[T]
	RemoveListener(listener obsfx2.InvalidationListener) listExpressionHelper[T]
	AddListListener(listener ListListener[T]) listExpressionHelper[T]
	RemoveListListener(listener ListListener[T]) listExpressionHelper[T]
}

type genericInvalidationListExpressionHelper[T any] struct {
	invalidationListeners obsfx2.HasListenersBase[obsfx2.InvalidationListener]
	listListeners         obsfx2.HasListenersBase[ListListener[T]]
}

func newGenericInvalidationListExpressionHelper[T any](invalidationListeners []obsfx2.InvalidationListener, listListeners []ListListener[T]) *genericInvalidationListExpressionHelper[T] {
	g := &genericInvalidationListExpressionHelper[T]{}

	for _, l := range invalidationListeners {
		g.AddListener(l)
	}

	for _, l := range listListeners {
		g.AddListListener(l)
	}

	return g
}

func (g *genericInvalidationListExpressionHelper[T]) OnListChanged(ev ListChangeEvent[T]) {
	g.invalidationListeners.ForEachListener(func(l obsfx2.InvalidationListener) bool {
		l.OnInvalidated(ev.List())

		return true
	})

	g.listListeners.ForEachListener(func(l ListListener[T]) bool {
		l.OnListChanged(ev)

		return true
	})
}

func (g *genericInvalidationListExpressionHelper[T]) AddListener(listener obsfx2.InvalidationListener) listExpressionHelper[T] {
	g.invalidationListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationListExpressionHelper[T]) RemoveListener(listener obsfx2.InvalidationListener) listExpressionHelper[T] {
	g.invalidationListeners.RemoveListener(listener)

	return g
}

func (g *genericInvalidationListExpressionHelper[T]) AddListListener(listener ListListener[T]) listExpressionHelper[T] {
	g.listListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationListExpressionHelper[T]) RemoveListListener(listener ListListener[T]) listExpressionHelper[T] {
	g.listListeners.RemoveListener(listener)

	return g
}

type singleInvalidationListExpressionHelper[T any] struct {
	listener obsfx2.InvalidationListener
}

func (s *singleInvalidationListExpressionHelper[T]) AddListListener(listener ListListener[T]) listExpressionHelper[T] {
	return newGenericInvalidationListExpressionHelper[T]([]obsfx2.InvalidationListener{s.listener}, []ListListener[T]{listener})
}

func (s *singleInvalidationListExpressionHelper[T]) RemoveListListener(listener ListListener[T]) listExpressionHelper[T] {
	return s
}

func (s *singleInvalidationListExpressionHelper[T]) AddListener(listener obsfx2.InvalidationListener) listExpressionHelper[T] {
	if s.listener == listener {
		return s
	}

	return newGenericInvalidationListExpressionHelper[T]([]obsfx2.InvalidationListener{s.listener, listener}, nil)
}

func (s *singleInvalidationListExpressionHelper[T]) RemoveListener(listener obsfx2.InvalidationListener) listExpressionHelper[T] {
	if s.listener == listener {
		return nil
	}

	return s
}

func (s *singleInvalidationListExpressionHelper[T]) OnListChanged(ev ListChangeEvent[T]) {
	s.listener.OnInvalidated(ev.List())
}

type singleListListenerExpressionHelper[T any] struct {
	listener ListListener[T]
}

func (s *singleListListenerExpressionHelper[T]) OnListChanged(ev ListChangeEvent[T]) {
	s.listener.OnListChanged(ev)
}

func (s *singleListListenerExpressionHelper[T]) AddListener(listener obsfx2.InvalidationListener) listExpressionHelper[T] {
	return newGenericInvalidationListExpressionHelper[T]([]obsfx2.InvalidationListener{listener}, []ListListener[T]{s.listener})
}

func (s *singleListListenerExpressionHelper[T]) RemoveListener(listener obsfx2.InvalidationListener) listExpressionHelper[T] {
	return s
}

func (s *singleListListenerExpressionHelper[T]) AddListListener(listener ListListener[T]) listExpressionHelper[T] {
	if reflect.DeepEqual(s.listener, listener) {
		return s
	}

	return newGenericInvalidationListExpressionHelper[T](nil, []ListListener[T]{s.listener, listener})
}

func (s *singleListListenerExpressionHelper[T]) RemoveListListener(listener ListListener[T]) listExpressionHelper[T] {
	if s.listener == listener {
		return nil
	}

	return s
}
