package introspectfx

import (
	"reflect"
)

type Value interface {
	Go() reflect.Value
	Type() Type
}

type nestedValue struct {
	receiver Value
	prop     Property
}

func (n nestedValue) Go() reflect.Value {
	return n.prop.GetRawValue(n.receiver)
}

func (n nestedValue) Type() Type {
	return n.prop.Type()
}
