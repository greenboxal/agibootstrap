package typesystem

import (
	"reflect"
	"sync"
	"time"

	"github.com/invopop/jsonschema"
)

func newTypeSystem() *typeSystem {
	ts := &typeSystem{
		typesByName: map[string]Type{},
		typesByType: map[reflect.Type]Type{},

		globalJsonSchema: jsonschema.Schema{
			Definitions: map[string]*jsonschema.Schema{},
		},
	}

	return ts
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

func (ts *typeSystem) registerTypeWithOptions(typ reflect.Type, opts ...typeCreationOption) Type {
	t := newTypeFromReflection(typ, opts...)

	ts.Register(t)

	return t
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

func (ts *typeSystem) Initialize() {
	ts.registerTypeWithOptions(jsonSchemaType, func(t *basicType) {
		t.decodeFromAny = func(v Value) (Value, error) {
			if v.checkTypeDataKind(reflect.Bool) {
				if v.v.Bool() {
					return ValueOf(*jsonschema.TrueSchema), nil
				} else {
					return ValueOf(*jsonschema.FalseSchema), nil
				}
			}

			return Value{}, nil
		}
	})

	ts.registerTypeWithOptions(timeType, func(t *basicType) {
		t.decodeFromAny = func(v Value) (Value, error) {
			result := time.Time{}

			if !v.v.IsValid() || !v.v.CanInterface() {
				return ValueOf(time.Time{}), nil
			}

			switch val := v.v.Interface().(type) {
			case time.Time:
				result = val
			case string:
				var err error

				result, err = time.Parse(time.RFC3339, val)

				if err != nil {
					return Value{}, err
				}
			case int64:
				result = time.Unix(0, val)
			case uint64:
				result = time.Unix(0, int64(val))
			case float64:
				result = time.Unix(0, int64(val))
			}

			return ValueOf(result), nil
		}
	})
}
