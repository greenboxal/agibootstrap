package graphfx

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type GraphExpressionHelper[K comparable, N Node, E Edge] interface {
	GraphListener[K, N, E]

	AddListener(listener obsfx.InvalidationListener) GraphExpressionHelper[K, N, E]
	RemoveListener(listener obsfx.InvalidationListener) GraphExpressionHelper[K, N, E]
	AddGraphListener(listener GraphListener[K, N, E]) GraphExpressionHelper[K, N, E]
	RemoveGraphListener(listener GraphListener[K, N, E]) GraphExpressionHelper[K, N, E]
}

type genericInvalidationGraphExpressionHelper[K comparable, N Node, E Edge] struct {
	invalidationListeners obsfx.HasListenersBase[obsfx.InvalidationListener]
	graphListeners        obsfx.HasListenersBase[GraphListener[K, N, E]]
}

func newGenericInvalidationGraphExpressionHelper[K comparable, N Node, E Edge](invalidationListeners []obsfx.InvalidationListener, listListeners []GraphListener[K, N, E]) *genericInvalidationGraphExpressionHelper[K, N, E] {
	g := &genericInvalidationGraphExpressionHelper[K, N, E]{}

	for _, l := range invalidationListeners {
		g.AddListener(l)
	}

	for _, l := range listListeners {
		g.AddGraphListener(l)
	}

	return g
}

func (g *genericInvalidationGraphExpressionHelper[K, N, E]) OnGraphChanged(ev GraphChangeEvent[K, N, E]) {
	g.invalidationListeners.ForEachListener(func(l obsfx.InvalidationListener) bool {
		l.OnInvalidated(ev.Graph())

		return true
	})

	g.graphListeners.ForEachListener(func(l GraphListener[K, N, E]) bool {
		l.OnGraphChanged(ev)

		return true
	})
}

func (g *genericInvalidationGraphExpressionHelper[K, N, E]) AddListener(listener obsfx.InvalidationListener) GraphExpressionHelper[K, N, E] {
	g.invalidationListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationGraphExpressionHelper[K, N, E]) RemoveListener(listener obsfx.InvalidationListener) GraphExpressionHelper[K, N, E] {
	g.invalidationListeners.RemoveListener(listener)

	return g
}

func (g *genericInvalidationGraphExpressionHelper[K, N, E]) AddGraphListener(listener GraphListener[K, N, E]) GraphExpressionHelper[K, N, E] {
	g.graphListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationGraphExpressionHelper[K, N, E]) RemoveGraphListener(listener GraphListener[K, N, E]) GraphExpressionHelper[K, N, E] {
	g.graphListeners.RemoveListener(listener)

	return g
}

type singleInvalidationGraphExpressionHelper[K comparable, N Node, E Edge] struct {
	listener obsfx.InvalidationListener
}

func (s *singleInvalidationGraphExpressionHelper[K, N, E]) AddGraphListener(listener GraphListener[K, N, E]) GraphExpressionHelper[K, N, E] {
	return newGenericInvalidationGraphExpressionHelper[K, N, E]([]obsfx.InvalidationListener{s.listener}, []GraphListener[K, N, E]{listener})
}

func (s *singleInvalidationGraphExpressionHelper[K, N, E]) RemoveGraphListener(listener GraphListener[K, N, E]) GraphExpressionHelper[K, N, E] {
	return s
}

func (s *singleInvalidationGraphExpressionHelper[K, N, E]) AddListener(listener obsfx.InvalidationListener) GraphExpressionHelper[K, N, E] {
	if s.listener == listener {
		return s
	}

	return newGenericInvalidationGraphExpressionHelper[K, N, E]([]obsfx.InvalidationListener{s.listener, listener}, nil)
}

func (s *singleInvalidationGraphExpressionHelper[K, N, E]) RemoveListener(listener obsfx.InvalidationListener) GraphExpressionHelper[K, N, E] {
	if s.listener == listener {
		return nil
	}

	return s
}

func (s *singleInvalidationGraphExpressionHelper[K, N, E]) OnGraphChanged(ev GraphChangeEvent[K, N, E]) {
	s.listener.OnInvalidated(ev.Graph())
}

type singleGraphListenerExpressionHelper[K comparable, N Node, E Edge] struct {
	listener GraphListener[K, N, E]
}

func (s *singleGraphListenerExpressionHelper[K, N, E]) OnGraphChanged(ev GraphChangeEvent[K, N, E]) {
	s.listener.OnGraphChanged(ev)
}

func (s *singleGraphListenerExpressionHelper[K, N, E]) AddListener(listener obsfx.InvalidationListener) GraphExpressionHelper[K, N, E] {
	return newGenericInvalidationGraphExpressionHelper[K, N, E]([]obsfx.InvalidationListener{listener}, []GraphListener[K, N, E]{s.listener})
}

func (s *singleGraphListenerExpressionHelper[K, N, E]) RemoveListener(listener obsfx.InvalidationListener) GraphExpressionHelper[K, N, E] {
	return s
}

func (s *singleGraphListenerExpressionHelper[K, N, E]) AddGraphListener(listener GraphListener[K, N, E]) GraphExpressionHelper[K, N, E] {
	if s.listener == listener {
		return s
	}

	return newGenericInvalidationGraphExpressionHelper[K, N, E](nil, []GraphListener[K, N, E]{s.listener, listener})
}

func (s *singleGraphListenerExpressionHelper[K, N, E]) RemoveGraphListener(listener GraphListener[K, N, E]) GraphExpressionHelper[K, N, E] {
	if s.listener == listener {
		return nil
	}

	return s
}
