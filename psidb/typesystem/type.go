package typesystem

import (
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/schema"
)

type typeInitializer interface {
	initialize(ts TypeSystem)
}

type basicType struct {
	self Type
	name TypeName

	primitiveKind PrimitiveKind
	runtimeType   reflect.Type

	ipldType               schema.Type
	ipldPrimitive          ipld.NodePrototype
	ipldPrototype          schema.TypedPrototype
	ipldRepresentationKind datamodel.Kind
	jsonSchema             jsonschema.Schema

	decorations []Decoration

	operators   []Operator
	operatorMap []map[string]Operator

	decodeFromAny func(v Value) (Value, error)

	universe TypeSystem
}

func (bt *basicType) initialize(ts TypeSystem) {
	bt.universe = ts
	bt.operatorMap = make([]map[string]Operator, len(bt.operators))

	bt.jsonSchema.Description = ts.(*typeSystem).LookupComment(bt.runtimeType, "")
}

func (bt *basicType) Name() TypeName                         { return bt.name }
func (bt *basicType) PrimitiveKind() PrimitiveKind           { return bt.primitiveKind }
func (bt *basicType) RuntimeType() reflect.Type              { return bt.runtimeType }
func (bt *basicType) Struct() StructType                     { return bt.self.(StructType) }
func (bt *basicType) List() ListType                         { return bt.self.(ListType) }
func (bt *basicType) Map() MapType                           { return bt.self.(MapType) }
func (bt *basicType) IpldType() schema.Type                  { return bt.ipldType }
func (bt *basicType) IpldPrimitive() ipld.NodePrototype      { return bt.ipldPrimitive }
func (bt *basicType) IpldPrototype() schema.TypedPrototype   { return bt.ipldPrototype }
func (bt *basicType) IpldRepresentationKind() datamodel.Kind { return bt.ipldRepresentationKind }
func (bt *basicType) JsonSchema() *jsonschema.Schema         { return &bt.jsonSchema }

func (bt *basicType) AssignableTo(other Type) bool {
	if other == nil {
		return false
	}

	return bt.runtimeType.AssignableTo(other.RuntimeType())
}

func (bt *basicType) ConvertFromAny(v Value) (Value, error) {
	if bt.decodeFromAny != nil {
		result, err := bt.decodeFromAny(v)

		if err != nil {
			return Value{}, err
		}

		if result.IsValid() {
			return result, nil
		}
	}

	return Value{}, fmt.Errorf("cannot convert from any to %s", bt.name)
}

type interfaceType struct {
	basicType
}

type structType struct {
	basicType

	fields   []Field
	fieldMap map[string]Field
}

func (st *structType) NumField() int                { return len(st.fields) }
func (st *structType) Field(name string) Field      { return st.fieldMap[name] }
func (st *structType) FieldByIndex(index int) Field { return st.fields[index] }

type listType struct {
	basicType

	elem Type
}

func (lt *listType) Elem() Type { return lt.elem }

type mapType struct {
	basicType

	key Type
	val Type
}

func (mt *mapType) Key() Type   { return mt.key }
func (mt *mapType) Value() Type { return mt.val }
