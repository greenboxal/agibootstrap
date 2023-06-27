package introspectfx

import (
	"reflect"
	"sync"
)

type Type interface {
	Name() string
	RuntimeType() reflect.Type

	Property(name string) Property
	Properties() []Property
}
type TypeBuilder struct {
	sync.Mutex
	m sync.Mutex

	name        string
	runtimeType reflect.Type

	properties map[string]Property

	dirty bool
}

func NewTypeBuilder(name string, runtimeType reflect.Type) *TypeBuilder {
	return &TypeBuilder{
		name:        name,
		runtimeType: runtimeType,

		properties: map[string]Property{},
	}
}

func (tb *TypeBuilder) Name() string {
	return tb.name
}

func (tb *TypeBuilder) RuntimeType() reflect.Type {
	return tb.runtimeType
}

func (tb *TypeBuilder) Property(name string) Property {
	return tb.properties[name]
}

func (tb *TypeBuilder) Properties() []Property {
	props := make([]Property, 0, len(tb.properties))

	for _, p := range tb.properties {
		props = append(props, p)
	}

	return props
}

func (tb *TypeBuilder) WithProperties(props ...Property) *TypeBuilder {
	for _, prop := range props {
		tb.WithProperty(prop)
	}

	return tb
}

func (tb *TypeBuilder) WithProperty(prop Property) *TypeBuilder {
	tb.m.Lock()
	defer tb.m.Unlock()

	// FIXME: Shouldn't allow
	// if _, ok := tb.properties[prop.Name()]; ok {
	// 	panic("property already exists")
	// }

	tb.properties[prop.Name()] = prop
	tb.dirty = true

	return tb
}

func (tb *TypeBuilder) IsDirty() bool {
	return tb.dirty
}

func (tb *TypeBuilder) Build() Type {
	tb.dirty = false

	props := make([]Property, 0, len(tb.properties))

	for _, p := range tb.properties {
		props = append(props, p)
	}

	return newReflectedType(tb.name, tb.runtimeType, props)
}
