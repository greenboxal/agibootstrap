package collectionsfx

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/obsfx"
)

type mapExpressionHelper[K comparable, V any] interface {
	MapListener[K, V]

	AddListener(listener obsfx.InvalidationListener) mapExpressionHelper[K, V]
	RemoveListener(listener obsfx.InvalidationListener) mapExpressionHelper[K, V]
	AddMapListener(listener MapListener[K, V]) mapExpressionHelper[K, V]
	RemoveMapListener(listener MapListener[K, V]) mapExpressionHelper[K, V]
}

type genericInvalidationMapExpressionHelper[K comparable, V any] struct {
	invalidationListeners obsfx.HasListenersBase[obsfx.InvalidationListener]
	mapListeners          obsfx.HasListenersBase[MapListener[K, V]]
}

func newGenericInvalidationMapExpressionHelper[K comparable, V any](invalidationListeners []obsfx.InvalidationListener, listListeners []MapListener[K, V]) *genericInvalidationMapExpressionHelper[K, V] {
	g := &genericInvalidationMapExpressionHelper[K, V]{}

	for _, l := range invalidationListeners {
		g.AddListener(l)
	}

	for _, l := range listListeners {
		g.AddMapListener(l)
	}

	return g
}

func (g *genericInvalidationMapExpressionHelper[K, V]) OnMapChanged(ev MapChangeEvent[K, V]) {
	g.invalidationListeners.ForEachListener(func(l obsfx.InvalidationListener) bool {
		l.OnInvalidated(ev.Map)

		return true
	})

	g.mapListeners.ForEachListener(func(l MapListener[K, V]) bool {
		l.OnMapChanged(ev)

		return true
	})
}

func (g *genericInvalidationMapExpressionHelper[K, V]) AddListener(listener obsfx.InvalidationListener) mapExpressionHelper[K, V] {
	g.invalidationListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationMapExpressionHelper[K, V]) RemoveListener(listener obsfx.InvalidationListener) mapExpressionHelper[K, V] {
	g.invalidationListeners.RemoveListener(listener)

	return g
}

func (g *genericInvalidationMapExpressionHelper[K, V]) AddMapListener(listener MapListener[K, V]) mapExpressionHelper[K, V] {
	g.mapListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationMapExpressionHelper[K, V]) RemoveMapListener(listener MapListener[K, V]) mapExpressionHelper[K, V] {
	g.mapListeners.RemoveListener(listener)

	return g
}

type singleInvalidationMapExpressionHelper[K comparable, T any] struct {
	listener obsfx.InvalidationListener
}

func (s *singleInvalidationMapExpressionHelper[K, V]) AddMapListener(listener MapListener[K, V]) mapExpressionHelper[K, V] {
	return newGenericInvalidationMapExpressionHelper[K, V]([]obsfx.InvalidationListener{s.listener}, []MapListener[K, V]{listener})
}

func (s *singleInvalidationMapExpressionHelper[K, V]) RemoveMapListener(listener MapListener[K, V]) mapExpressionHelper[K, V] {
	return s
}

func (s *singleInvalidationMapExpressionHelper[K, V]) AddListener(listener obsfx.InvalidationListener) mapExpressionHelper[K, V] {
	if s.listener == listener {
		return s
	}

	return newGenericInvalidationMapExpressionHelper[K, V]([]obsfx.InvalidationListener{s.listener, listener}, nil)
}

func (s *singleInvalidationMapExpressionHelper[K, V]) RemoveListener(listener obsfx.InvalidationListener) mapExpressionHelper[K, V] {
	if s.listener == listener {
		return nil
	}

	return s
}

func (s *singleInvalidationMapExpressionHelper[K, V]) OnMapChanged(ev MapChangeEvent[K, V]) {
	s.listener.OnInvalidated(ev.Map)
}

type singleMapListenerExpressionHelper[K comparable, V any] struct {
	listener MapListener[K, V]
}

func (s *singleMapListenerExpressionHelper[K, V]) OnMapChanged(ev MapChangeEvent[K, V]) {
	s.listener.OnMapChanged(ev)
}

func (s *singleMapListenerExpressionHelper[K, V]) AddListener(listener obsfx.InvalidationListener) mapExpressionHelper[K, V] {
	return newGenericInvalidationMapExpressionHelper[K, V]([]obsfx.InvalidationListener{listener}, []MapListener[K, V]{s.listener})
}

func (s *singleMapListenerExpressionHelper[K, V]) RemoveListener(listener obsfx.InvalidationListener) mapExpressionHelper[K, V] {
	return s
}

func (s *singleMapListenerExpressionHelper[K, V]) AddMapListener(listener MapListener[K, V]) mapExpressionHelper[K, V] {
	if s.listener == listener {
		return s
	}

	return newGenericInvalidationMapExpressionHelper[K, V](nil, []MapListener[K, V]{s.listener, listener})
}

func (s *singleMapListenerExpressionHelper[K, V]) RemoveMapListener(listener MapListener[K, V]) mapExpressionHelper[K, V] {
	if s.listener == listener {
		return nil
	}

	return s
}
