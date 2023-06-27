package introspectfx

import (
	"reflect"
)

type reflectedType struct {
	name      string
	typ       reflect.Type
	props     map[string]Property
	propSlice []Property
}

func newReflectedType(name string, typ reflect.Type, props []Property) Type {
	propMap := make(map[string]Property, len(props))

	for _, p := range props {
		propMap[p.Name()] = p
	}

	return &reflectedType{
		name:      name,
		typ:       typ,
		props:     propMap,
		propSlice: props,
	}
}

func (r *reflectedType) Name() string {
	return r.name
}

func (r *reflectedType) RuntimeType() reflect.Type {
	return r.typ
}

func (r *reflectedType) Property(name string) Property {
	return r.props[name]
}

func (r *reflectedType) Properties() []Property {
	return r.propSlice
}

func adjustPointers(expected reflect.Type, value reflect.Value) reflect.Value {
	if expected.Kind() != value.Kind() {
		if expected.Kind() == reflect.Ptr && value.Kind() != reflect.Ptr {
			value = value.Addr()
		} else if expected.Kind() != reflect.Ptr && value.Kind() == reflect.Ptr {
			for value.Kind() == reflect.Ptr {
				value = value.Elem()
			}
		}
	}

	return value
}
