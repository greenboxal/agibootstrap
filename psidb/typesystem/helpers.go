package typesystem

import (
	"reflect"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/schema"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
)

var globalTypeSystem = newTypeSystem()

func init() {
	globalTypeSystem.Initialize()
}

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
	if v, ok := v.(Value); ok {
		return v
	}

	if v, ok := v.(reflect.Value); ok {
		return ValueFrom(v)
	}

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

func MakeList(t reflect.Type, length, cap int) Value {
	sliceType := reflect.SliceOf(t)

	v := Value{
		v: reflect.MakeSlice(sliceType, length, cap),
	}

	v.typ = TypeFrom(v.v.Type())

	return v
}

func MakeMap(k, v reflect.Type, length int) Value {
	rt := reflect.MapOf(k, v)
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

func FlattenJsonSchema(u TypeSystem, schema *jsonschema.Schema) *jsonschema.Schema {
	var refStack []*jsonschema.Schema
	var flatten func(schema *jsonschema.Schema, clone bool) *jsonschema.Schema

	ts := u.(*typeSystem)

	cloneAndQueue := func(schema *jsonschema.Schema) *jsonschema.Schema {
		if schema == nil {
			return nil
		}

		return flatten(schema, true)
	}

	flatten = func(originalSchema *jsonschema.Schema, clone bool) *jsonschema.Schema {
		if slices.Contains(refStack, originalSchema) {
			panic("circular ref")
		}

		refStack = append(refStack, originalSchema)

		defer func() {
			last := len(refStack) - 1

			if refStack[last] != originalSchema {
				panic("invalid ref stack")
			}

			refStack = refStack[:last]
		}()

		resultSchema := originalSchema

		if clone {
			cloned := *resultSchema
			resultSchema = &cloned
		}

		if resultSchema.Ref != "" {
			ref := ts.LookupByJsonSchemaRef(resultSchema.Ref)

			if ref == nil {
				panic("invalid ref")
			}

			*resultSchema = *ref
		}

		if resultSchema.Properties != nil {
			props := orderedmap.New()

			for _, key := range resultSchema.Properties.Keys() {
				v, _ := resultSchema.Properties.Get(key)
				prop := v.(*jsonschema.Schema)

				props.Set(key, cloneAndQueue(prop))
			}

			resultSchema.Properties = props
		}

		if resultSchema.Items != nil {
			resultSchema.Items = cloneAndQueue(resultSchema.Items)
		}

		if resultSchema.AdditionalProperties != nil {
			resultSchema.AdditionalProperties = cloneAndQueue(resultSchema.AdditionalProperties)
		}

		if resultSchema.AllOf != nil {
			resultSchema.AllOf = lo.Map(resultSchema.AllOf, func(item *jsonschema.Schema, _ int) *jsonschema.Schema {
				return cloneAndQueue(item)
			})
		}

		if resultSchema.AnyOf != nil {
			resultSchema.AnyOf = lo.Map(resultSchema.AnyOf, func(item *jsonschema.Schema, _ int) *jsonschema.Schema {
				return cloneAndQueue(item)
			})
		}

		if resultSchema.OneOf != nil {
			resultSchema.OneOf = lo.Map(resultSchema.OneOf, func(item *jsonschema.Schema, _ int) *jsonschema.Schema {
				return cloneAndQueue(item)
			})
		}

		return resultSchema
	}

	return cloneAndQueue(schema)
}
