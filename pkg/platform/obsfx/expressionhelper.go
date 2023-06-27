package obsfx

import (
	"reflect"
)

type ExpressionHelper[T any] interface {
	InvalidationListener

	AddListener(listener InvalidationListener) ExpressionHelper[T]
	RemoveListener(listener InvalidationListener) ExpressionHelper[T]

	AddChangeListener(listener ChangeListener[T]) ExpressionHelper[T]
	RemoveChangeListener(listener ChangeListener[T]) ExpressionHelper[T]
}

type genericInvalidationExpressionHelper[T any] struct {
	invalidationListeners HasListenersBase[InvalidationListener]
	changeListeners       HasListenersBase[ChangeListener[T]]
	currentValue          T
}

func newGenericInvalidationExpressionHelper[T any](invalidationListeners []InvalidationListener, changeListeners []ChangeListener[T]) *genericInvalidationExpressionHelper[T] {
	g := &genericInvalidationExpressionHelper[T]{}

	for _, l := range invalidationListeners {
		g.AddListener(l)
	}

	for _, l := range changeListeners {
		g.AddChangeListener(l)
	}

	return g
}

func (g *genericInvalidationExpressionHelper[T]) OnInvalidated(observable Observable) {
	g.invalidationListeners.ForEachListener(func(l InvalidationListener) bool {
		l.OnInvalidated(observable)

		return true
	})

	if len(g.changeListeners.listeners) > 0 {
		obs := observable.(ObservableValue[T])
		previous := g.currentValue
		val := obs.Value()

		if reflect.DeepEqual(previous, val) {
			return
		}

		g.currentValue = val

		g.changeListeners.ForEachListener(func(l ChangeListener[T]) bool {
			l.OnChanged(obs, previous, val)

			return true
		})
	}
}

func (g *genericInvalidationExpressionHelper[T]) AddListener(listener InvalidationListener) ExpressionHelper[T] {
	g.invalidationListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationExpressionHelper[T]) RemoveListener(listener InvalidationListener) ExpressionHelper[T] {
	g.invalidationListeners.RemoveListener(listener)

	return g
}

func (g *genericInvalidationExpressionHelper[T]) AddChangeListener(listener ChangeListener[T]) ExpressionHelper[T] {
	g.changeListeners.AddListener(listener)

	return g
}

func (g *genericInvalidationExpressionHelper[T]) RemoveChangeListener(listener ChangeListener[T]) ExpressionHelper[T] {
	g.changeListeners.RemoveListener(listener)

	return g
}

type singleInvalidationExpressionHelper[T any] struct {
	listener InvalidationListener
}

func NewSingleInvalidationExpressionHelper[T any](listener InvalidationListener) ExpressionHelper[T] {
	return &singleInvalidationExpressionHelper[T]{listener: listener}
}

func (s *singleInvalidationExpressionHelper[T]) AddChangeListener(listener ChangeListener[T]) ExpressionHelper[T] {
	return newGenericInvalidationExpressionHelper[T]([]InvalidationListener{s.listener}, []ChangeListener[T]{listener})
}

func (s *singleInvalidationExpressionHelper[T]) RemoveChangeListener(listener ChangeListener[T]) ExpressionHelper[T] {
	return s
}

func (s *singleInvalidationExpressionHelper[T]) AddListener(listener InvalidationListener) ExpressionHelper[T] {
	if reflect.DeepEqual(s.listener, listener) {
		return s
	}

	return newGenericInvalidationExpressionHelper[T]([]InvalidationListener{s.listener, listener}, nil)
}

func (s *singleInvalidationExpressionHelper[T]) RemoveListener(listener InvalidationListener) ExpressionHelper[T] {
	if reflect.DeepEqual(s.listener, listener) {
		return nil
	}

	return s
}

func (s *singleInvalidationExpressionHelper[T]) OnInvalidated(obs Observable) {
	s.listener.OnInvalidated(obs)
}

type singleChangeListenerExpressionHelper[T any] struct {
	listener     ChangeListener[T]
	currentValue T
}

func NewSingleChangeListenerExpressionHelper[T any](listener ChangeListener[T]) ExpressionHelper[T] {
	return &singleChangeListenerExpressionHelper[T]{listener: listener}
}

func (s *singleChangeListenerExpressionHelper[T]) OnInvalidated(observable Observable) {
	obs := observable.(ObservableValue[T])
	previous := s.currentValue
	val := obs.Value()

	if reflect.DeepEqual(previous, val) {
		return
	}

	s.currentValue = val

	s.listener.OnChanged(obs, previous, val)
}

func (s *singleChangeListenerExpressionHelper[T]) AddListener(listener InvalidationListener) ExpressionHelper[T] {
	return newGenericInvalidationExpressionHelper[T]([]InvalidationListener{listener}, []ChangeListener[T]{s.listener})
}

func (s *singleChangeListenerExpressionHelper[T]) RemoveListener(listener InvalidationListener) ExpressionHelper[T] {
	return s
}

func (s *singleChangeListenerExpressionHelper[T]) AddChangeListener(listener ChangeListener[T]) ExpressionHelper[T] {
	if reflect.DeepEqual(s.listener, listener) {
		return s
	}

	return newGenericInvalidationExpressionHelper[T](nil, []ChangeListener[T]{s.listener, listener})
}

func (s *singleChangeListenerExpressionHelper[T]) RemoveChangeListener(listener ChangeListener[T]) ExpressionHelper[T] {
	if reflect.DeepEqual(s.listener, listener) {
		return nil
	}

	return s
}
