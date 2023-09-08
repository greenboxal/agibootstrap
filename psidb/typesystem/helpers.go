package typesystem

import (
	"reflect"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/schema"
)

var globalTypeSystem = newTypeSystem()

func Universe() TypeSystem {
	return globalTypeSystem
}

func GetType[T any]() Type {
	return TypeFrom(reflect.TypeOf((*T)(nil)).Elem())
}

func TypeOf(v interface{}) Type {
	return Universe().LookupByType(reflect.TypeOf(v))
}

func TypeFrom(t reflect.Type) Type {
	return Universe().LookupByType(t)
}

func ValueOf(v interface{}) Value {
	return Value{
		typ: TypeOf(v),
		v:   reflect.ValueOf(v),
	}
}

func ValueFrom(v reflect.Value) Value {
	return Value{
		typ: TypeFrom(v.Type()),
		v:   v,
	}
}

func New(t Type) Value {
	return Value{
		typ: t,
		v:   reflect.New(t.RuntimeType()).Elem(),
	}
}

func MakeList(t Type, length, cap int) Value {
	sliceType := reflect.SliceOf(t.RuntimeType())

	v := Value{
		v: reflect.MakeSlice(sliceType, length, cap),
	}

	v.typ = TypeFrom(v.v.Type())

	return v
}

func MakeMap(k, v Type, length int) Value {
	rt := reflect.MapOf(k.RuntimeType(), v.RuntimeType())
	t := TypeFrom(rt)

	return Value{
		typ: t,
		v:   reflect.MakeMapWithSize(rt, length),
	}
}

func Wrap(v any) ipld.Node {
	if n, ok := v.(ipld.Node); ok {
		return n
	}

	return ValueOf(v).AsNode().(schema.TypedNode)
}

func Unwrap(v ipld.Node) any {
	res, ok := TryUnwrap[any](v)

	if !ok {
		return v
	}

	return res
}

func TryUnwrap[T any](v ipld.Node) (def T, ok bool) {
	val, ok := v.(valueNode)

	if !ok {
		return
	}

	return TryCast[T](val.v.Indirect())
}

func TryCast[T any](v reflect.Value) (def T, ok bool) {
	t := reflect.TypeOf((*T)(nil)).Elem()

	if !v.IsValid() {
		return def, false
	}

	if v.Kind() == reflect.Interface && v.IsNil() {
		return def, false
	}

	if v.Kind() == reflect.Pointer && v.IsNil() {
		return def, false
	}

	if v.CanConvert(t) {
		return v.Convert(t).Interface().(T), true
	}

	if v.CanInterface() {
		r, ok := v.Interface().(T)

		if ok {
			return r, true
		}
	}

	if v.Kind() != reflect.Ptr && v.CanAddr() {
		v = v.Addr()

		if v.CanConvert(t) {
			return v.Convert(t).Interface().(T), true
		}

		if v.CanInterface() {
			r, ok := v.Interface().(T)

			if ok {
				return r, true
			}
		}
	} else if v.Kind() == reflect.Ptr {
		for v.Kind() == reflect.Ptr {
			v = v.Elem()

			if v.CanConvert(t) {
				return v.Convert(t).Interface().(T), true
			}

			if v.CanInterface() {
				r, ok := v.Interface().(T)

				if ok {
					return r, true
				}
			}
		}
	}

	return def, false
}

func Implements[T any](typ reflect.Type) bool {
	iface := reflect.TypeOf((*T)(nil)).Elem()

	if typ.Implements(iface) {
		return true
	}

	if typ.Kind() != reflect.Ptr {
		typ = reflect.PtrTo(typ)
	}

	return typ.Implements(iface)
}
