package typesystem

import (
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/schema"
)

type Type interface {
	Name() TypeName
	PrimitiveKind() PrimitiveKind
	RuntimeType() reflect.Type

	IpldType() schema.Type
	IpldPrimitive() ipld.NodePrototype
	IpldPrototype() schema.TypedPrototype
	IpldRepresentationKind() datamodel.Kind
	JsonSchema() *jsonschema.Schema

	Struct() StructType
	List() ListType
	Map() MapType

	AssignableTo(other Type) bool
}

type StructType interface {
	Type

	NumField() int
	Field(name string) Field
	FieldByIndex(index int) Field
}

type ListType interface {
	Type

	Elem() Type
}

type MapType interface {
	Type

	Key() Type
	Value() Type
}

type Field interface {
	Name() string
	Type() Type
	DeclaringType() StructType
	IsVirtual() bool
	IsNullable() bool
	IsOptional() bool

	Resolve(v Value) Value
}

func Universe() *TypeSystem {
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
