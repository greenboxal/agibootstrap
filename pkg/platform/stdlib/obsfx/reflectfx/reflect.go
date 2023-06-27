package reflectfx

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	obsfx2 "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
)

type ReflectedProperty interface {
	Name() string
	Type() reflect.Type

	Get(receiver any) obsfx2.IProperty
}

type NodeType interface {
	Type() reflect.Type
	ObservableProperties() []ReflectedProperty
	GetProperty(name string) ReflectedProperty
}

type reflectionBasedProperty struct {
	name  string
	typ   reflect.Type
	field reflect.StructField
}

func (r *reflectionBasedProperty) Name() string {
	return r.name
}

func (r *reflectionBasedProperty) Type() reflect.Type {
	return r.typ
}

func (r *reflectionBasedProperty) Get(receiver any) obsfx2.IProperty {
	v := reflect.ValueOf(receiver)
	p := v

	for _, i := range r.field.Index {
		p = v.Field(i)
	}

	if p.Kind() != reflect.Pointer {
		p = p.Addr()
	}

	return p.Interface().(obsfx2.IProperty)
}

type ReflectedNodeType struct {
	typ   reflect.Type
	props map[string]ReflectedProperty
}

func (r ReflectedNodeType) Type() reflect.Type {
	return r.typ
}

func (r ReflectedNodeType) GetProperty(name string) ReflectedProperty {
	return r.props[name]
}

func (r ReflectedNodeType) ObservableProperties() []ReflectedProperty {
	props := make([]ReflectedProperty, 0, len(r.props))

	for _, prop := range props {
		props = append(props, prop)
	}

	return props
}

var propertyType = reflect.TypeOf((*obsfx2.IProperty)(nil)).Elem()
var observableType = reflect.TypeOf((*obsfx2.Observable)(nil)).Elem()
var basicObservableListType = reflect.TypeOf((*collectionsfx.BasicObservableList)(nil)).Elem()
var reflectedTypeCache = map[string]NodeType{}
var reflectedTypeCacheMutex sync.RWMutex

func IsObservable(typ reflect.Type) bool {
	return typ.Implements(observableType)
}

func IsObservableList(typ reflect.Type) bool {
	return typ.Implements(basicObservableListType)
}

func IsObservableMap(typ reflect.Type) bool {
	_, ok := typ.MethodByName("AddMapListener")

	return IsObservable(typ) && ok
}

func IsObservableSet(typ reflect.Type) bool {
	_, ok := typ.MethodByName("AddSetListener")

	return IsObservable(typ) && ok
}

func IsProperty(typ reflect.Type) bool {
	if typ.Kind() != reflect.Pointer {
		typ = reflect.PointerTo(typ)
	}

	return typ.Implements(propertyType)
}

func AsObservable[T any](val reflect.Value) (obsfx2.ObservableValue[T], bool) {
	if val.Kind() != reflect.Pointer {
		if !val.CanAddr() {
			return nil, false
		}

		val = val.Addr()
	}

	r, ok := val.Interface().(obsfx2.ObservableValue[T])

	return r, ok
}

func GetPropertyType(typ reflect.Type) reflect.Type {
	valueMethod, hasValueMethod := typ.MethodByName("Value")

	if !hasValueMethod {
		panic("type has no Value method")
	}

	return valueMethod.Type.Out(0)
}

func Reflect(value interface{}) NodeType {
	return ReflectValue(reflect.ValueOf(value))
}

func ReflectValue(value reflect.Value) NodeType {
	return ReflectType(value.Type())
}

func ReflectType(typ reflect.Type) NodeType {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	id := fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())

	reflectedTypeCacheMutex.Lock()
	defer reflectedTypeCacheMutex.Unlock()

	existing := reflectedTypeCache[id]

	if existing == nil {
		existing = NewReflectedNodeType(typ)

		reflectedTypeCache[id] = existing
	}

	return existing
}

func NewReflectedNodeType(typ reflect.Type) NodeType {
	var scanFields func(typ reflect.Type, indexBase []int)

	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	props := map[string]ReflectedProperty{}

	scanFields = func(typ reflect.Type, indexBase []int) {
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			t := f.Type

			f.Index = append(indexBase, f.Index...)

			if f.Anonymous && t.Kind() == reflect.Struct {
				scanFields(t, f.Index)
				continue
			}

			if t.Kind() != reflect.Pointer {
				t = reflect.PointerTo(t)
			}

			if !t.Implements(propertyType) {
				continue
			}

			valueMethod, hasValueMethod := typ.MethodByName("Value")

			if !hasValueMethod {
				continue
			}

			propType := valueMethod.Type.Out(0)
			name := f.Name

			if trimmed := strings.TrimSuffix(name, "Property"); trimmed != "" {
				name = trimmed
			}

			prop := &reflectionBasedProperty{
				name:  name,
				typ:   propType,
				field: f,
			}

			props[prop.name] = prop
		}
	}

	scanFields(typ, nil)

	return &ReflectedNodeType{
		typ:   typ,
		props: props,
	}
}
