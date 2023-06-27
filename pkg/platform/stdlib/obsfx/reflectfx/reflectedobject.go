package reflectfx

import (
	"reflect"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
)

type value struct {
	typ NodeType
	val interface{}
}

func (v value) Type() reflect.Type {
	return v.typ.Type()
}

func (v value) GetProperty(name string) (obsfx.IProperty, bool) {
	prop := v.typ.GetProperty(name)

	if prop == nil {
		return nil, false
	}

	return prop.Get(v.val), true
}

func (v value) GetProperties() []obsfx.IProperty {
	props := v.typ.ObservableProperties()
	result := make([]obsfx.IProperty, len(props))

	for i, p := range props {
		result[i] = p.Get(v.val)
	}

	return result
}

type Value interface {
	Type() reflect.Type

	GetProperty(name string) (obsfx.IProperty, bool)
	GetProperties() []obsfx.IProperty
}

func ValueOf(v interface{}) Value {
	typ := Reflect(v)

	return &value{
		typ: typ,
		val: v,
	}
}
