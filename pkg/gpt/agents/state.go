package agents

import "github.com/greenboxal/agibootstrap/pkg/psi"

type WorldStateKey[T any] string

func (k WorldStateKey[T]) String() string {
	return string(k)
}

type WorldState interface {
	psi.Node

	Get(key string) any
	Set(key string, value any)
}

func GetState[T any](state WorldState, k WorldStateKey[T]) (def T) {
	v := state.Get(k.String())

	if v == nil {
		return def
	}

	return v.(T)
}

func SetState[T any](state WorldState, k WorldStateKey[T], v T) {
	state.Set(k.String(), v)
}
