package typesystem

import (
	"reflect"
	"sync"

	"github.com/invopop/jsonschema"
)

func newTypeSystem() TypeSystem {
	return &typeSystem{
		typesByName: map[string]Type{},
		typesByType: map[reflect.Type]Type{},

		globalJsonSchema: jsonschema.Schema{
			Definitions: map[string]*jsonschema.Schema{},
		},
	}
}

type typeSystem struct {
	m sync.RWMutex

	typesByName map[string]Type
	typesByType map[reflect.Type]Type

	globalJsonSchema jsonschema.Schema
}

func (ts *typeSystem) GlobalJsonSchema() *jsonschema.Schema {
	return &ts.globalJsonSchema
}

func (ts *typeSystem) Register(t Type) {
	register := func() bool {
		ts.m.Lock()
		defer ts.m.Unlock()

		return ts.doRegister(t)
	}

	if register() {
		if init, ok := t.(typeInitializer); ok {
			init.initialize(ts)
		}
	}
}

func (ts *typeSystem) doRegister(t Type) bool {
	name := t.Name().NormalizedFullNameWithArguments()

	if _, ok := ts.typesByType[t.RuntimeType()]; ok {
		panic("type already registered")
	}

	if _, ok := ts.typesByName[name]; ok {
		return false //panic("type already registered")
	}

	ts.typesByType[t.RuntimeType()] = t
	ts.typesByName[name] = t
	ts.globalJsonSchema.Definitions[t.Name().NormalizedFullNameWithArguments()] = t.JsonSchema()

	return true
}

func (ts *typeSystem) LookupByName(name string) Type {
	ts.m.RLock()
	defer ts.m.RUnlock()

	return ts.typesByName[name]
}

func (ts *typeSystem) LookupByType(typ reflect.Type) Type {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	get := func() (Type, bool) {
		ts.m.Lock()
		defer ts.m.Unlock()

		if existing := ts.typesByType[typ]; existing != nil {
			return existing, false
		}

		newType := newTypeFromReflection(typ)

		ts.doRegister(newType)

		return newType, true
	}

	t, isNew := get()

	if isNew {
		if init, ok := t.(typeInitializer); ok {
			init.initialize(ts)
		}
	}

	return t
}
