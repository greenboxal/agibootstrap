package psi

import "fmt"

type EdgeReference interface {
	GetKind() EdgeKind
	GetName() string
	GetIndex() int64
	GetKey() EdgeKey
}

type EdgeKey struct {
	Kind  EdgeKind `json:"Kind"`
	Name  string   `json:"Name"`
	Index int64    `json:"Index"`
}

func (k EdgeKey) GetKind() EdgeKind { return k.Kind }
func (k EdgeKey) GetName() string   { return k.Name }
func (k EdgeKey) GetIndex() int64   { return k.Index }
func (k EdgeKey) GetKey() EdgeKey   { return k }
func (k EdgeKey) String() string    { return fmt.Sprintf("%s=%d:%s", k.Kind, k.Index, k.Name) }

type TypedEdgeReference[T Node] interface {
	EdgeReference

	GetTypedKind() TypedEdgeKind[T]
}

type TypedEdgeKey[T Node] struct {
	Kind  TypedEdgeKind[T] `json:"Kind"`
	Name  string           `json:"Name"`
	Index int64            `json:"Index"`
}

func (k TypedEdgeKey[T]) GetTypedKind() TypedEdgeKind[T] { return k.Kind }
func (k TypedEdgeKey[T]) GetKind() EdgeKind              { return EdgeKind(k.Kind) }
func (k TypedEdgeKey[T]) GetName() string                { return k.Name }
func (k TypedEdgeKey[T]) GetIndex() int64                { return k.Index }
func (k TypedEdgeKey[T]) String() string                 { return fmt.Sprintf("%s=%d:%s", k.Kind, k.Index, k.Name) }

func (k TypedEdgeKey[T]) GetKey() EdgeKey {
	return EdgeKey{
		Kind:  k.GetKind(),
		Name:  k.GetName(),
		Index: k.GetIndex(),
	}
}
