package typesystem

import (
	"encoding"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
	"golang.org/x/exp/slices"
)

var anyType = reflect.TypeOf((*any)(nil)).Elem()
var timeType = reflect.TypeOf((*time.Time)(nil)).Elem()
var durationType = reflect.TypeOf((*time.Duration)(nil)).Elem()
var jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
var jsonUnmarshalerType = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
var textMarshalerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
var binaryMarshalerType = reflect.TypeOf((*encoding.BinaryMarshaler)(nil)).Elem()
var binaryUnmarshalerType = reflect.TypeOf((*encoding.BinaryUnmarshaler)(nil)).Elem()
var jsonSchemaType = reflect.TypeOf((*jsonschema.Schema)(nil)).Elem()

type typeCreationOption func(t *basicType)

func newTypeFromReflection(typ reflect.Type, opts ...typeCreationOption) Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	primitiveKind := getPrimitiveKind(typ)

	switch primitiveKind {
	case PrimitiveKindInvalid:
		panic("invalid type")

	case PrimitiveKindList:
		return newListType(typ, opts...)

	case PrimitiveKindMap:
		return newMapType(typ, opts...)

	case PrimitiveKindStruct:
		return newStructType(typ, opts...)

	case PrimitiveKindInterface:
		return newInterfaceType(typ, opts...)

	case PrimitiveKindFunction:
		return newFunctionType(typ, opts...)

	default:
		return newScalarType(typ, opts...)
	}
}

func newInterfaceType(typ reflect.Type, option ...typeCreationOption) *interfaceType {
	it := &interfaceType{
		basicType: basicType{
			name:          GetTypeName(typ),
			primitiveKind: getPrimitiveKind(typ),
			runtimeType:   typ,
		},
	}

	it.self = it

	for _, opt := range option {
		opt(&it.basicType)
	}

	return it
}

func (it *interfaceType) initialize(ts TypeSystem) {
	it.basicType.initialize(ts)

	if Implements[ipld.Link](it.runtimeType) {
		it.ipldPrimitive = basicnode.Prototype.Link
		it.ipldRepresentationKind = datamodel.Kind_Link
	} else {
		it.ipldPrimitive = basicnode.Prototype.Any
		it.ipldRepresentationKind = datamodel.Kind_Map
	}

	it.ipldType = schema.SpawnAny(it.name.ToTitle())
	it.ipldPrototype = &ValuePrototype{T: it}

	it.jsonSchema.Type = "interface"
	it.jsonSchema.Properties = orderedmap.New()

	for _, m := range it.methods {
		it.jsonSchema.Properties.Set(m.Name(), &jsonschema.Schema{
			Ref: "#/$defs/" + mangleFunctionName(m.(*reflectedMethod).m.Type).NormalizedFullNameWithArguments(),
		})
	}
}

func newScalarType(typ reflect.Type, option ...typeCreationOption) *scalarType {
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
	case PrimitiveKindVector:
		st.ipldType = schema.SpawnList(st.name.ToTitle(), "Float64", false)
		st.ipldPrimitive = basicnode.Prototype.List
		st.ipldRepresentationKind = datamodel.Kind_List
		st.jsonSchema.Type = "array"
		st.jsonSchema.Items = &jsonschema.Schema{Type: "number"}
	default:
		panic("invalid scalar type")
	}

	st.self = st
	st.ipldPrototype = &ValuePrototype{T: st}

	for _, opt := range option {
		opt(&st.basicType)
	}

	return st
}

func newStructType(typ reflect.Type, option ...typeCreationOption) *structType {
	st := &structType{
		basicType: basicType{
			name:          GetTypeName(typ),
			primitiveKind: PrimitiveKindStruct,
			runtimeType:   typ,
		},

		fieldMap: map[string]Field{},
	}

	st.self = st

	st.decorations = DecorationsForType(typ)

	for _, dec := range st.decorations {
		switch dec := dec.(type) {
		case NameDecoration:
			st.name = ParseTypeName(dec.Name)
		}
	}

	for _, opt := range option {
		opt(&st.basicType)
	}

	return st
}

func (st *structType) initialize(ts TypeSystem) {
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
			tagParts := strings.Split(tag, ",")

			if hasTag {
				taggedName = tagParts[0]
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
					walkFields(t, patchedField.Index)
					continue
				}
			}

			nullable := directKind == reflect.Ptr || directKind == reflect.Interface
			optional := nullable || hasTag && slices.Contains(tagParts, "omitempty")

			fld := &reflectedField{
				fieldBase: fieldBase{
					declaringType: st,
					name:          name,
					typ:           ts.LookupByType(f.Type),
					nullable:      nullable,
					optional:      optional,
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
		if f.IsVirtual() {
			continue
		}

		ipldFields[i] = schema.SpawnStructField(
			f.Name(),
			f.Type().Name().ToTitle(),
			f.IsOptional(),
			f.IsNullable(),
		)

		refPath := "#/$defs/" + f.Type().Name().NormalizedFullNameWithArguments()
		schemaRef := &jsonschema.Schema{
			Ref: refPath,
		}

		if f.RuntimeField() != nil {
			schemaRef.Description = ts.(*typeSystem).LookupComment(st.RuntimeType(), f.RuntimeField().Name)

			if k := f.RuntimeField().Type.Kind(); k != reflect.Pointer && k != reflect.Interface {
				schemaRef = f.Type().JsonSchema()
			}

			tag := f.RuntimeField().Tag.Get("jsonschmea")
			tagOpts := strings.Split(tag, ",")
			for _, opt := range tagOpts {
				kv := strings.Split(opt, "=")

				switch kv[0] {
				case "title":
					schemaRef.Title = kv[1]
				case "description":
					schemaRef.Description = kv[1]
				case "type":
					schemaRef.Type = kv[1]
				case "format":
					schemaRef.Format = kv[1]
				case "pattern":
					schemaRef.Pattern = kv[1]
				case "minLength":
					schemaRef.MinLength, _ = strconv.Atoi(kv[1])
				case "maxLength":
					schemaRef.MaxLength, _ = strconv.Atoi(kv[1])
				case "minimum":
					schemaRef.Minimum, _ = strconv.Atoi(kv[1])
				case "maximum":
					schemaRef.Maximum, _ = strconv.Atoi(kv[1])
				case "exclusiveMinimum":
					schemaRef.ExclusiveMinimum = true
				case "exclusiveMaximum":
					schemaRef.ExclusiveMaximum = true
				}
			}
		}

		st.jsonSchema.Properties.Set(f.Name(), schemaRef)

		if !f.IsOptional() {
			st.jsonSchema.Required = append(st.jsonSchema.Required, f.Name())
		}
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
		st.ipldRepresentationKind = GetRepresentationKind(st.ipldType)

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

	st.ipldPrototype = &ValuePrototype{T: st}
}

func newMapType(typ reflect.Type, option ...typeCreationOption) *mapType {
	keyName := GetTypeName(typ.Key())
	valName := GetTypeName(typ.Elem())
	name := GetTypeName(typ).WithInParameters(keyName, valName)

	mt := &mapType{
		basicType: basicType{
			primitiveKind: PrimitiveKindMap,
			name:          name,
			runtimeType:   typ,
		},
	}

	mt.self = mt

	for _, opt := range option {
		opt(&mt.basicType)
	}

	return mt
}

func (mt *mapType) initialize(ts TypeSystem) {
	mt.basicType.initialize(ts)

	typ := mt.runtimeType

	mt.key = ts.LookupByType(typ.Key())
	mt.val = ts.LookupByType(typ.Elem())

	if mt.key == nil {
		panic("key type not found")
	}

	if mt.val == nil {
		panic("value type not found")
	}

	mt.ipldType = schema.SpawnMap(mt.name.ToTitle(), mt.key.Name().ToTitle(), mt.val.Name().ToTitle(), false)
	mt.ipldPrimitive = basicnode.Prototype.Map
	mt.ipldPrototype = &ValuePrototype{T: mt}
	mt.ipldRepresentationKind = datamodel.Kind_Map

	mt.jsonSchema.Type = "object"

	mt.jsonSchema.PatternProperties = map[string]*jsonschema.Schema{
		".*": {
			Ref: "#/$defs/" + mt.Value().Name().NormalizedFullNameWithArguments(),
		},
	}
}

func newListType(typ reflect.Type, option ...typeCreationOption) *listType {
	switch typ.Kind() {
	case reflect.Array:

	}

	valName := GetTypeName(typ.Elem())
	name := GetTypeName(typ).WithInParameters(valName)

	lt := &listType{
		basicType: basicType{
			primitiveKind: PrimitiveKindList,
			name:          name,
			runtimeType:   typ,
		},
	}

	lt.self = lt

	for _, opt := range option {
		opt(&lt.basicType)
	}

	return lt
}

func (lt *listType) initialize(ts TypeSystem) {
	lt.basicType.initialize(ts)

	typ := lt.runtimeType

	lt.elem = ts.LookupByType(typ.Elem())
	lt.ipldType = schema.SpawnList(lt.name.ToTitle(), lt.elem.Name().ToTitle(), false)
	lt.ipldPrimitive = basicnode.Prototype.List
	lt.ipldPrototype = &ValuePrototype{T: lt}
	lt.ipldRepresentationKind = datamodel.Kind_List

	refPath := "#/$defs/" + lt.Elem().Name().NormalizedFullNameWithArguments()
	schemaRef := &jsonschema.Schema{
		Ref: refPath,
	}

	lt.jsonSchema.Type = "array"
	lt.jsonSchema.Items = schemaRef
}

func mangleFunctionName(typ reflect.Type) TypeName {
	in := make([]TypeName, typ.NumIn())
	out := make([]TypeName, typ.NumOut())

	inCountArg := TypeName{}
	inCountArg.Name = strconv.FormatInt(int64(typ.NumIn()), 10)

	for i := 0; i < typ.NumIn(); i++ {
		in[i] = GetTypeName(typ.In(i))
	}

	for i := 0; i < typ.NumOut(); i++ {
		out[i] = GetTypeName(typ.Out(i))
	}

	return GetTypeName(typ).WithInParameters(in...).WithOutParameters(out...)
}

type functionType struct {
	basicType

	in  []Type
	out []Type
}

func newFunctionType(typ reflect.Type, option ...typeCreationOption) *functionType {
	name := mangleFunctionName(typ)

	ft := &functionType{
		basicType: basicType{
			primitiveKind: PrimitiveKindFunction,
			name:          name,
			runtimeType:   typ,
		},
	}

	ft.self = ft

	for _, opt := range option {
		opt(&ft.basicType)
	}

	return ft
}

func makeSchemaRef(t Type) *jsonschema.Schema {
	return &jsonschema.Schema{
		Ref: "#/$defs/" + t.Name().NormalizedFullNameWithArguments(),
	}
}

func (ft *functionType) initialize(ts TypeSystem) {
	ft.basicType.initialize(ts)

	ft.in = make([]Type, 0, ft.runtimeType.NumIn())
	ft.out = make([]Type, 0, ft.runtimeType.NumOut())

	for i := 0; i < ft.runtimeType.NumIn(); i++ {
		ft.in = append(ft.in, ts.LookupByType(ft.runtimeType.In(i)))
	}

	for i := 0; i < ft.runtimeType.NumOut(); i++ {
		ft.out = append(ft.out, ts.LookupByType(ft.runtimeType.Out(i)))
	}

	ft.ipldType = schema.SpawnAny(ft.name.ToTitle())
	ft.ipldPrimitive = basicnode.Prototype.Any
	ft.ipldPrototype = &ValuePrototype{T: ft}
	ft.ipldRepresentationKind = datamodel.Kind_Link

	ft.jsonSchema.Type = "function"
	ft.jsonSchema.Properties = orderedmap.New()

	requestType := &jsonschema.Schema{}

	if len(ft.in) == 1 {
		requestType = makeSchemaRef(ft.in[0])
	} else {
		requestType.Type = "array"
		requestType.PrefixItems = make([]*jsonschema.Schema, 0, len(ft.in))
		for _, in := range ft.in {
			requestType.PrefixItems = append(requestType.PrefixItems, makeSchemaRef(in))
		}
	}

	responseType := &jsonschema.Schema{}

	if len(ft.out) == 1 {
		responseType = makeSchemaRef(ft.out[0])
	} else {
		responseType.Type = "array"
		responseType.PrefixItems = make([]*jsonschema.Schema, 0, len(ft.out))
		for _, out := range ft.out {
			responseType.PrefixItems = append(responseType.PrefixItems, makeSchemaRef(out))
		}
	}

	ft.jsonSchema.Properties.Set("request_type", requestType)
	ft.jsonSchema.Properties.Set("response_type", responseType)
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
		fallthrough
	case reflect.Uintptr:
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
	case reflect.Complex64:
		return PrimitiveKindVector
	case reflect.Complex128:
		return PrimitiveKindVector
	case reflect.Map:
		return PrimitiveKindMap
	case reflect.Struct:
		return PrimitiveKindStruct
	case reflect.Chan:
		fallthrough
	case reflect.UnsafePointer:
		fallthrough
	case reflect.Interface:
		return PrimitiveKindInterface
	case reflect.Func:
		return PrimitiveKindFunction
	case reflect.Pointer:
		return getPrimitiveKind(typ.Elem())
	default:
		panic("not supported")
	}
}
