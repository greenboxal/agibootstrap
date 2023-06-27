package introspectfx

import "reflect"

func IntrospectValue(value Value) Node {
	return newValueNode(value)
}

func Introspect(value interface{}) Node {
	return newValueNode(ValueOf(value))
}

func ValueOf(val any) Value {
	return ValueFor(reflect.ValueOf(val))
}

func ValueFor(val reflect.Value) Value {
	return value{
		val: val,
		typ: TypeFor(val.Type()),
	}
}

func TypeOf(val any) Type {
	return TypeFor(reflect.TypeOf(val))
}

func TypeFor(t reflect.Type) Type {
	return globalTypeCache.TypeOf(t)
}

type value struct {
	val reflect.Value
	typ Type
}

func (v value) Go() reflect.Value {
	return v.val
}

func (v value) Type() Type {
	return v.typ
}
