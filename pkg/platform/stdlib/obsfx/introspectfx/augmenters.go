package introspectfx

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/reflectfx"
)

type TypeAugmenter interface {
	AugmentType(tb *TypeBuilder) error
}

type TypeAugmenterFunc func(tb *TypeBuilder) error

func (t TypeAugmenterFunc) AugmentType(tb *TypeBuilder) error {
	return t(tb)
}

func RegisterProperty(typ Type, prop Property) {
	globalTypeCache.RegisterProperty(typ, prop)
}

func RegisterPropertyMethod[T any](name string) {
	typ := TypeFor(reflect.TypeOf((*T)(nil)).Elem())

	method, ok := typ.RuntimeType().MethodByName(name)

	if !ok {
		method, ok = reflect.PointerTo(typ.RuntimeType()).MethodByName(name)

		if !ok {
			panic(fmt.Errorf("method %s not found", name))
		}
	}

	prop := &methodProperty{}
	prop.name = name
	prop.runtimeType = method.Type.Out(0)
	prop.method = method

	RegisterProperty(typ, prop)
}

func RegisterPropertyFunc[T, V any](name string, fn func(T) V) {
	typ := TypeFor(reflect.TypeOf((*T)(nil)).Elem())

	prop := &funcProperty{}
	prop.name = name
	prop.runtimeType = reflect.TypeOf((*V)(nil)).Elem()
	prop.fn = reflect.ValueOf(fn)

	RegisterProperty(typ, prop)
}

func RegisterAugmenter(aug TypeAugmenter) {
	globalTypeCache.RegisterAugmenter(aug)
}

func init() {
	RegisterAugmenter(TypeAugmenterFunc(AugmentObservablesByReflection))
}

func AugmentObservablesByReflection(tb *TypeBuilder) error {
	typ := tb.RuntimeType()

	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if field.Type.Kind() == reflect.Func {
			continue
		}

		name := field.Name

		tag, tagOk := field.Tag.Lookup("introspect")

		//if !field.IsExported() && !tagOk {
		//	continue
		//}

		if tagOk && tag == "-" {
			continue
		}

		if !field.IsExported() {
			name = "__" + name
		}

		if reflectfx.IsProperty(field.Type) {
			if strings.HasSuffix(name, "Property") && name != "Property" {
				name = strings.TrimSuffix(name, "Property")
			}
		}

		if tb.Property(name) != nil {
			continue
		}

		prop := newReflectedFieldProperty(name, field)

		tb.WithProperty(prop)
	}

	return nil
}
