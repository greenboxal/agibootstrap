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

	ConvertFromAny(v Value) (Value, error)

	NumMethods() int
	Method(name string) Method
	MethodByIndex(index int) Method
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

	RuntimeField() *reflect.StructField
}

type Method interface {
	Name() string
	DeclaringType() Type

	NumIn() int
	In(index int) Type

	NumOut() int
	Out(index int) Type

	IsVariadic() bool

	Call(receiver Value, args ...Value) ([]Value, error)
	CallSlice(receiver Value, args ...Value) ([]Value, error)
}

type TypeSystem interface {
	GlobalJsonSchema() any

	Register(t Type)
	LookupByName(name string) Type
	LookupByType(typ reflect.Type) Type

	LookupComment(t reflect.Type, name string) string

	CompileBundleFor(schema *jsonschema.Schema) *jsonschema.Schema
	CompileFlatBundleFor(schema *jsonschema.Schema) *jsonschema.Schema
}
