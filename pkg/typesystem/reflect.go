package typesystem

import (
	"encoding"
	"reflect"
	"strings"

	`github.com/iancoleman/orderedmap`
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
	"golang.org/x/exp/slices"
)

func newTypeFromReflection(typ reflect.Type) Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	primitiveKind := getPrimitiveKind(typ)

	switch primitiveKind {
	case PrimitiveKindInvalid:
		panic("invalid type")

	case PrimitiveKindList:
		return newListType(typ)

	case PrimitiveKindMap:
		return newMapType(typ)

	case PrimitiveKindStruct:
		return newStructType(typ)

	case PrimitiveKindInterface:
		return newInterfaceType(typ)

	default:
		return newScalarType(typ)
	}
}

func newInterfaceType(typ reflect.Type) *interfaceType {
	it := &interfaceType{
		basicType: basicType{
			name:          GetTypeName(typ),
			primitiveKind: getPrimitiveKind(typ),
			runtimeType:   typ,
		},
	}

	it.self = it

	return it
}

func (it *interfaceType) initialize(ts *TypeSystem) {
	it.basicType.initialize(ts)

	if Implements[ipld.Link](it.runtimeType) {
		it.ipldPrimitive = basicnode.Prototype.Link
		it.ipldRepresentationKind = datamodel.Kind_Link
	} else {
		it.ipldPrimitive = basicnode.Prototype.Any
		it.ipldRepresentationKind = datamodel.Kind_Map
	}

	it.ipldType = schema.SpawnAny(it.name.ToTitle())
	it.ipldPrototype = &valuePrototype{typ: it}
}

func newScalarType(typ reflect.Type) *scalarType {
	st := &scalarType{
		basicType: basicType{
			name:          GetTypeName(typ),
			primitiveKind: getPrimitiveKind(typ),
			runtimeType:   typ,
		},
	}

	switch st.PrimitiveKind() {
	case PrimitiveKindBoolean:
		st.ipldType = schema.SpawnBool(st.name.ToTitle())
		st.ipldPrimitive = basicnode.Prototype.Bool
		st.ipldRepresentationKind = datamodel.Kind_Bool
		st.jsonSchema.Type = "boolean"
	case PrimitiveKindString:
		st.ipldType = schema.SpawnString(st.name.ToTitle())
		st.ipldPrimitive = basicnode.Prototype.String
		st.ipldRepresentationKind = datamodel.Kind_String
		st.jsonSchema.Type = "string"
	case PrimitiveKindFloat:
		st.ipldType = schema.SpawnFloat(st.name.ToTitle())
		st.ipldPrimitive = basicnode.Prototype.Float
		st.ipldRepresentationKind = datamodel.Kind_Float
		st.jsonSchema.Type = "number"
	case PrimitiveKindInt:
		fallthrough
	case PrimitiveKindUnsignedInt:
		st.ipldType = schema.SpawnInt(st.name.ToTitle())
		st.ipldPrimitive = basicnode.Prototype.Int
		st.ipldRepresentationKind = datamodel.Kind_Int
		st.jsonSchema.Type = "number"
	case PrimitiveKindBytes:
		st.ipldType = schema.SpawnBytes(st.name.ToTitle())
		st.ipldPrimitive = basicnode.Prototype.Bytes
		st.ipldRepresentationKind = datamodel.Kind_Bytes
		st.jsonSchema.Type = "string"
	case PrimitiveKindLink:
		if Implements[TypedLink](typ) {
			elemTyp := reflect.New(typ).Interface().(TypedLink).LinkedObjectType()

			st.ipldType = schema.SpawnLinkReference(st.name.ToTitle(), elemTyp.Name().ToTitle())
		} else {
			st.ipldType = schema.SpawnLink(st.name.ToTitle())
		}

		st.ipldPrimitive = basicnode.Prototype.Link
		st.ipldRepresentationKind = datamodel.Kind_Link
		st.jsonSchema.Type = "string"
	default:
		panic("invalid scalar type")
	}

	st.self = st
	st.ipldPrototype = &valuePrototype{typ: st}

	return st
}

func newStructType(typ reflect.Type) *structType {
	st := &structType{
		basicType: basicType{
			name:          GetTypeName(typ),
			primitiveKind: PrimitiveKindStruct,
			runtimeType:   typ,
		},

		fieldMap: map[string]Field{},
	}

	st.self = st

	return st
}

func (st *structType) initialize(ts *TypeSystem) {
	var walkFields func(typ reflect.Type, indexBase []int)

	st.basicType.initialize(ts)

	typ := st.runtimeType

	st.fields = make([]Field, 0, typ.NumField())

	walkFields = func(typ reflect.Type, indexBase []int) {
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)

			if !f.IsExported() {
				continue
			}

			directType := f.Type

			for directType.Kind() == reflect.Ptr {
				directType = directType.Elem()
			}

			directKind := directType.Kind()

			if directKind == reflect.Func || directKind == reflect.Chan || directKind == reflect.UnsafePointer {
				continue
			}

			patchedField := f
			patchedField.Index = nil
			patchedField.Index = append(patchedField.Index, indexBase...)
			patchedField.Index = append(patchedField.Index, f.Index...)

			name := f.Name
			taggedName := ""
			tag, hasTag := f.Tag.Lookup("json")

			if hasTag {
				parts := strings.Split(tag, ",")
				taggedName = parts[0]
			}

			if taggedName == "-" {
				continue
			} else if taggedName != "" {
				name = taggedName
			}

			if f.Anonymous && taggedName == "" {
				t := f.Type

				for t.Kind() == reflect.Ptr {
					t = t.Elem()
				}

				if t.Kind() == reflect.Struct {
					walkFields(f.Type, patchedField.Index)
					continue
				}
			}

			fld := &reflectedField{
				fieldBase: fieldBase{
					declaringType: st,
					name:          name,
					typ:           ts.LookupByType(f.Type),
				},

				runtimeField: patchedField,
			}

			st.fields = append(st.fields, fld)
			st.fieldMap[fld.name] = fld
		}
	}

	walkFields(typ, nil)

	ipldFields := make([]schema.StructField, len(st.fields))

	st.jsonSchema.Properties = orderedmap.New()

	for i, f := range st.fields {
		k := f.Type().RuntimeType().Kind()
		nullable := k == reflect.Ptr || k == reflect.Interface

		ipldFields[i] = schema.SpawnStructField(
			f.Name(),
			f.Type().IpldType().Name(),
			nullable,
			nullable,
		)

		st.jsonSchema.Properties.Set(f.Name(), f.Type().JsonSchema())
		st.jsonSchema.Required = append(st.jsonSchema.Required, f.Name())
	}

	var repr schema.StructRepresentation

	if Implements[encoding.TextMarshaler](typ) {
		st.ipldPrimitive = basicnode.Prototype.String
		st.ipldRepresentationKind = datamodel.Kind_String
		st.ipldType = schema.SpawnString(st.name.ToTitle())
		st.jsonSchema.Type = "string"
	} else if Implements[encoding.BinaryMarshaler](typ) {
		st.ipldPrimitive = basicnode.Prototype.Bytes
		st.ipldRepresentationKind = datamodel.Kind_Bytes
		st.ipldType = schema.SpawnBytes(st.name.ToTitle())
		st.jsonSchema.Type = "string"
	} else if typ.NumField() == 1 {
		f := typ.Field(0)
		tag := f.Tag.Get("ipld")
		parts := strings.Split(tag+",", ",")
		options := parts[1:]

		isInline := slices.Contains(options, "inline")

		if f.Anonymous && isInline {
			repr = schema.SpawnStructRepresentationStringjoin("/")

			st.ipldPrimitive = basicnode.Prototype.String
			st.ipldRepresentationKind = datamodel.Kind_String
			st.ipldType = schema.SpawnString(st.name.ToTitle())
			st.jsonSchema.Type = "string"
		}
	}

	if repr == nil {
		repr = schema.SpawnStructRepresentationMap(map[string]string{})
	}

	if st.ipldType == nil {
		st.ipldType = schema.SpawnStruct(st.name.ToTitle(), ipldFields, repr)
	}

	if st.ipldPrimitive == nil {
		st.ipldRepresentationKind = getReprKind(st.ipldType)

		switch st.ipldRepresentationKind {
		case datamodel.Kind_Map:
			st.ipldPrimitive = basicnode.Prototype.Map
			st.jsonSchema.Type = "object"
		case datamodel.Kind_List:
			st.ipldPrimitive = basicnode.Prototype.List
			st.jsonSchema.Type = "array"
		case datamodel.Kind_String:
			st.ipldPrimitive = basicnode.Prototype.String
			st.jsonSchema.Type = "string"
		case datamodel.Kind_Bytes:
			st.ipldPrimitive = basicnode.Prototype.Bytes
			st.jsonSchema.Type = "string"
		case datamodel.Kind_Bool:
			st.ipldPrimitive = basicnode.Prototype.Bool
			st.jsonSchema.Type = "boolean"
		case datamodel.Kind_Int:
			st.ipldPrimitive = basicnode.Prototype.Int
			st.jsonSchema.Type = "number"
		case datamodel.Kind_Float:
			st.ipldPrimitive = basicnode.Prototype.Float
			st.jsonSchema.Type = "number"
		}
	}

	st.ipldPrototype = &valuePrototype{typ: st}
}

func newMapType(typ reflect.Type) *mapType {
	keyName := GetTypeName(typ.Key())
	valName := GetTypeName(typ.Elem())
	name := GetTypeName(typ).WithParameters(keyName, valName)

	mt := &mapType{
		basicType: basicType{
			primitiveKind: PrimitiveKindMap,
			name:          name,
			runtimeType:   typ,
		},
	}

	mt.self = mt

	return mt
}

func (mt *mapType) initialize(ts *TypeSystem) {
	mt.basicType.initialize(ts)

	typ := mt.runtimeType

	mt.key = ts.LookupByType(typ.Key())
	mt.val = ts.LookupByType(typ.Elem())

	mt.ipldType = schema.SpawnMap(mt.name.ToTitle(), mt.key.IpldType().Name(), mt.val.IpldType().Name(), false)
	mt.ipldPrimitive = basicnode.Prototype.Map
	mt.ipldPrototype = &valuePrototype{typ: mt}
	mt.ipldRepresentationKind = datamodel.Kind_Map
	mt.jsonSchema.Type = "object"
}

func newListType(typ reflect.Type) *listType {
	valName := GetTypeName(typ.Elem())
	name := GetTypeName(typ).WithParameters(valName)

	lt := &listType{
		basicType: basicType{
			primitiveKind: PrimitiveKindList,
			name:          name,
			runtimeType:   typ,
		},
	}

	lt.self = lt

	return lt
}

func (lt *listType) initialize(ts *TypeSystem) {
	lt.basicType.initialize(ts)

	typ := lt.runtimeType

	lt.elem = ts.LookupByType(typ.Elem())
	lt.ipldType = schema.SpawnList(lt.name.ToTitle(), lt.elem.IpldType().Name(), false)
	lt.ipldPrimitive = basicnode.Prototype.List
	lt.ipldPrototype = &valuePrototype{typ: lt}
	lt.ipldRepresentationKind = datamodel.Kind_List
	lt.jsonSchema.Type = "array"
	lt.jsonSchema.Items = lt.elem.JsonSchema()
}

type HasIpldPrimitiveType interface {
	IpldPrimitiveType() PrimitiveKind
}

func getPrimitiveKind(typ reflect.Type) PrimitiveKind {
	if Implements[HasIpldPrimitiveType](typ) {
		return reflect.New(typ).Interface().(HasIpldPrimitiveType).IpldPrimitiveType()
	}

	if Implements[ipld.Link](typ) {
		return PrimitiveKindLink
	}

	switch typ.Kind() {
	case reflect.Invalid:
		return PrimitiveKindInvalid
	case reflect.Bool:
		return PrimitiveKindBoolean
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		return PrimitiveKindInt
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		return PrimitiveKindUnsignedInt
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		return PrimitiveKindFloat
	case reflect.String:
		return PrimitiveKindString
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		return PrimitiveKindList
	case reflect.Map:
		return PrimitiveKindMap
	case reflect.Struct:
		return PrimitiveKindStruct
	case reflect.Interface:
		return PrimitiveKindInterface
	case reflect.Pointer:
		return getPrimitiveKind(typ.Elem())
	default:
		panic("not supported")
	}
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

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
