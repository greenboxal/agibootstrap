package psi

import (
	"reflect"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
)

type AttributeKey interface {
	Name() string

	Type() typesystem.Type
	RuntimeType() reflect.Type

	String() string
}

type TypedAttributeKey[T any] interface {
	AttributeKey
}

func DefineAttribute[T any](name string) AttributeKey {
	return NewAttributeKey(name, typesystem.TypeFrom(reflect.TypeOf((*T)(nil)).Elem()))
}

func NewAttributeKey(name string, typ typesystem.Type) AttributeKey {
	return &attributeKey{
		name: name,
		typ:  typ,
	}
}

type attributeKey struct {
	name string
	typ  typesystem.Type
}

func (a *attributeKey) Name() string              { return a.name }
func (a *attributeKey) Type() typesystem.Type     { return a.typ }
func (a *attributeKey) RuntimeType() reflect.Type { return a.typ.RuntimeType() }
func (a *attributeKey) String() string            { return a.Name() }

func (a *attributeKey) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.name + `"`), nil
}

func (a *attributeKey) UnmarshalJSON(data []byte) error {
	a.name = string(data[1 : len(data)-1])

	return nil
}

func (a *attributeKey) MarshalText() ([]byte, error) {
	return []byte(a.name), nil
}

func (a *attributeKey) UnmarshalText(data []byte) error {
	a.name = string(data)

	return nil
}

func (a *attributeKey) MarshalBinary() ([]byte, error) {
	return []byte(a.name), nil
}

func (a *attributeKey) UnmarshalBinary(data []byte) error {
	a.name = string(data)

	return nil
}
