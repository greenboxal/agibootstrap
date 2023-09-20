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

	methods   []Method
	methodMap map[string]Method

	decodeFromAny func(v Value) (Value, error)

	universe TypeSystem
}

func (bt *basicType) initialize(ts TypeSystem) {
	bt.universe = ts
	bt.operatorMap = make([]map[string]Operator, len(bt.operators))
	bt.methodMap = make(map[string]Method, len(bt.methods))

	walkMethods := func(t reflect.Type) {
		for i := 0; i < t.NumMethod(); i++ {
			m := t.Method(i)
			mt := ts.LookupByType(m.Type)

			if !m.IsExported() {
				continue
			}

			reth := newReflectedMethod(bt.self, m, mt)

			bt.methods = append(bt.methods, reth)
		}
	}

	walkMethods(bt.runtimeType)

	if bt.runtimeType.Kind() != reflect.Pointer {
		walkMethods(reflect.PointerTo(bt.runtimeType))
	}

	// Thanks GPT
	/*for i, op := range bt.operators {
		if bt.operatorMap[op.Arity()] == nil {
			bt.operatorMap[op.Arity()] = make(map[string]Operator)
		}

		bt.operatorMap[op.Arity()][op.Name()] = op
	}*/

	for _, m := range bt.methods {
		bt.methodMap[m.Name()] = m
	}

	//bt.jsonSchema.ID = bt.jsonSchema.ID.Add(bt.name.NormalizedFullNameWithArguments())
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

func (lt *basicType) NumMethods() int           { return len(lt.methods) }
func (lt *basicType) Method(name string) Method { return lt.methodMap[name] }
func (lt *basicType) MethodByIndex(index int) Method {
	if index < 0 || index >= len(lt.methods) {
		panic("index out of range")
	}

	return lt.methods[index]
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
