package iterators

import "golang.org/x/exp/maps"

type KeyValue[K comparable, V any] struct {
	K K
	V V
}

func FromMap[K comparable, V any](m map[K]V) Iterator[KeyValue[K, V]] {
	return &mapEntryIterator[K, V]{m: m, keys: maps.Keys(m)}
}

func ToMap[K comparable, V any](iter Iterator[KeyValue[K, V]]) map[K]V {
	result := make(map[K]V)

	for iter.Next() {
		pair := iter.Value()
		result[pair.K] = pair.V
	}

	return result
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
