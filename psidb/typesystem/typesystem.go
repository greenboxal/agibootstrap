package typesystem

import (
	"encoding/json"
	"os"
	"path"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"github.com/jbenet/goprocess"
	"github.com/xeipuuv/gojsonpointer"
)

func newTypeSystem() *typeSystem {
	ts := &typeSystem{
		typesByName: map[string]Type{},
		typesByType: map[reflect.Type]Type{},

		globalJsonSchema: jsonschema.Schema{
			Definitions: map[string]*jsonschema.Schema{},
		},
	}

	ts.reflector.CommentMap = map[string]string{}

	if sdkPath := os.Getenv("PSIDB_SDK"); sdkPath != "" {
		if err := ExtractGoComments("github.com/greenboxal/agibootstrap/psidb", path.Join(sdkPath, "psidb"), ts.reflector.CommentMap); err != nil {
			panic(err)
		}
	}

	goprocess.Go(func(proc goprocess.Process) {
		for {
			select {
			case <-proc.Closing():
				return
			case <-time.After(time.Second * 5):
				ts.snapshotJsonSchema()
			}
		}
	})

	return ts
}
func (ts *typeSystem) CompileBundleFor(schema *jsonschema.Schema) *jsonschema.Schema {
	ts.m.RLock()
	defer ts.m.RUnlock()

	var bundle jsonschema.Schema

	bundle = *schema

	if bundle.Definitions == nil {
		bundle.Definitions = map[string]*jsonschema.Schema{}
	}

	var walk func(ref string, schema *jsonschema.Schema)

	seen := map[*jsonschema.Schema]bool{}

	walk = func(ref string, schema *jsonschema.Schema) {
		if ref == "" && schema.Ref != "" {
			ref = schema.Ref
		}

		if seen[schema] {
			return
		}

		seen[schema] = true

		if ref != "" {
			li := strings.LastIndex(ref, "/")
			name := ref[li+1:]

			def := ts.globalJsonSchema.Definitions[name]

			if def != nil {
				walk(ref, def)
			}

			bundle.Definitions[name] = def
		}

		if schema.Properties != nil && len(schema.Properties.Keys()) > 0 {
			for _, k := range schema.Properties.Keys() {
				v, _ := schema.Properties.Get(k)

				prop := v.(*jsonschema.Schema)

				walk("", prop)
			}
		}

		if schema.Items != nil {
			walk("", schema.Items)
		}
	}

	walk("", schema)

	return &bundle
}

func (ts *typeSystem) CompileFlatBundleFor(schema *jsonschema.Schema) *jsonschema.Schema {
	var bundle jsonschema.Schema
	var cloneAndPatch func(schema *jsonschema.Schema) *jsonschema.Schema

	ts.m.RLock()
	defer ts.m.RUnlock()

	bundle = *schema
	definitions := bundle.Definitions

	if definitions == nil {
		definitions = map[string]*jsonschema.Schema{}
	}

	cloneAndPatch = func(schema *jsonschema.Schema) *jsonschema.Schema {
		result := &jsonschema.Schema{}

		if schema.Ref != "" {
			li := strings.LastIndex(schema.Ref, "/")
			name := schema.Ref[li+1:]

			def := definitions[name]

			if def != nil {
				return def
			}

			schema = ts.LookupByJsonSchemaRef(schema.Ref)

			definitions[name] = schema
		}

		*result = *schema

		if schema.Properties != nil && len(schema.Properties.Keys()) > 0 {
			props := orderedmap.New()

			for _, k := range schema.Properties.Keys() {
				v, _ := schema.Properties.Get(k)
				prop := v.(*jsonschema.Schema)

				props.Set(k, cloneAndPatch(prop))
			}

			schema.Properties = props
		}

		if schema.PatternProperties != nil {
			props := map[string]*jsonschema.Schema{}

			for k, v := range schema.PatternProperties {
				props[k] = cloneAndPatch(v)
			}

			schema.PatternProperties = props
		}

		if schema.Items != nil {
			schema.Items = cloneAndPatch(schema.Items)
		}

		return result
	}

	bundle = *cloneAndPatch(schema)
	bundle.Definitions = definitions

	return &bundle
}

func (ts *typeSystem) LookupComment(t reflect.Type, name string) string {
	if ts.reflector.CommentMap == nil {
		return ""
	}

	n := t.PkgPath() + "." + t.Name()

	if name != "" {
		n = n + "." + name
	}

	return ts.reflector.CommentMap[n]
}

type typeSystem struct {
	m sync.RWMutex

	typesByName map[string]Type
	typesByType map[reflect.Type]Type

	reflector        jsonschema.Reflector
	globalJsonSchema jsonschema.Schema

	globalJsonSchemaSnapshot map[string]any
}

func (ts *typeSystem) GlobalJsonSchema() any {
	ts.m.RLock()
	defer ts.m.RUnlock()

	return ts.globalJsonSchemaSnapshot
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
	name := t.Name().MangledName()

	if _, ok := ts.typesByType[t.RuntimeType()]; ok {
		panic("type already registered")
	}

	if _, ok := ts.typesByName[name]; ok {
		return false //panic("type already registered")
	}

	ts.typesByType[t.RuntimeType()] = t
	ts.typesByName[name] = t
	ts.globalJsonSchema.Definitions[t.Name().MangledName()] = t.JsonSchema()

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

func (ts *typeSystem) LookupByJsonSchemaRef(ref string) *jsonschema.Schema {
	parts := strings.SplitN(ref, "#", 2)
	root := parts[0]
	pointer := ""

	if root != "" {
		panic("invalid ref")
	}

	if len(parts) > 1 {
		pointer = parts[1]
	}

	if pointer == "" {
		panic("invalid ref")
	}

	parsed, err := gojsonpointer.NewJsonPointer(parts[1])

	if err != nil {
		panic(err)
	}

	if ts.globalJsonSchemaSnapshot == nil {
		ts.snapshotJsonSchema()
	}

	result, _, err := parsed.Get(ts.globalJsonSchemaSnapshot)

	if err != nil {
		panic(err)
	}

	var resultSchema jsonschema.Schema

	data, err := json.Marshal(result)

	if err != nil {
		panic(err)
	}

	if err := resultSchema.UnmarshalJSON(data); err != nil {
		panic(err)
	}

	return &resultSchema
}

func (ts *typeSystem) snapshotJsonSchema() {
	ts.m.RLock()
	defer ts.m.RUnlock()

	data, err := ts.globalJsonSchema.MarshalJSON()

	if err != nil {
		panic(err)
	}

	var result map[string]any

	if err := json.Unmarshal(data, &result); err != nil {
		panic(err)
	}

	ts.globalJsonSchemaSnapshot = result
}
