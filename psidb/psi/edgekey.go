package psi

import (
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

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
func (k EdgeKey) String() string    { return k.AsPathElement().String() }

func (k EdgeKey) AsPathElement() PathElement {
	return PathElement{
		Kind:  k.Kind,
		Name:  k.Name,
		Index: k.Index,
	}
}

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
func (k TypedEdgeKey[T]) String() string                 { return k.AsPathElement().String() }

func (k TypedEdgeKey[T]) AsPathElement() PathElement {
	return PathElement{
		Kind:  k.Kind.Kind(),
		Name:  k.Name,
		Index: k.Index,
	}
}

func (k TypedEdgeKey[T]) GetKey() EdgeKey {
	return EdgeKey{
		Kind:  k.GetKind(),
		Name:  k.GetName(),
		Index: k.GetIndex(),
	}
}

func (k TypedEdgeKey[T]) IpldPrimitiveKind() typesystem.PrimitiveKind {
	return typesystem.PrimitiveKindString
}
func (k TypedEdgeKey[T]) MarshalText() ([]byte, error) { return []byte(k.String()), nil }
func (k *TypedEdgeKey[T]) UnmarshalText(text []byte) error {
	c, err := ParsePathElement(string(text))

	if err != nil {
		return err
	}

	k.Kind = TypedEdgeKind[T](c.Kind)
	k.Name = c.Name
	k.Index = c.Index

	return nil
}

func (k TypedEdgeKey[T]) MarshalBinary() ([]byte, error)     { return k.MarshalText() }
func (k *TypedEdgeKey[T]) UnmarshalBinary(data []byte) error { return k.UnmarshalText(data) }

func (k TypedEdgeKey[T]) MarshalJSON() ([]byte, error) { return []byte("\"" + k.String() + "\""), nil }
func (k *TypedEdgeKey[T]) UnmarshalJSON(data []byte) error {
	return k.UnmarshalText(data[1 : len(data)-1])
}

func (k EdgeKey) IpldPrimitiveKind() typesystem.PrimitiveKind { return typesystem.PrimitiveKindString }
func (k EdgeKey) MarshalText() ([]byte, error)                { return []byte(k.String()), nil }
func (k *EdgeKey) UnmarshalText(text []byte) error {
	c, err := ParsePathElement(string(text))

	if err != nil {
		return err
	}

	k.Kind = c.Kind
	k.Name = c.Name
	k.Index = c.Index

	return nil
}

func (k EdgeKey) MarshalBinary() ([]byte, error)     { return k.MarshalText() }
func (k *EdgeKey) UnmarshalBinary(data []byte) error { return k.UnmarshalText(data) }

func (k EdgeKey) MarshalJSON() ([]byte, error)     { return []byte("\"" + k.String() + "\""), nil }
func (k *EdgeKey) UnmarshalJSON(data []byte) error { return k.UnmarshalText(data[1 : len(data)-1]) }
