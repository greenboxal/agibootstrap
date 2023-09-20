package autoform

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-errors/errors"
	"github.com/ieee0824/go-deepmerge"
	"github.com/invopop/jsonschema"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/xeipuuv/gojsonpointer"
	"github.com/xeipuuv/gojsonreference"
	"github.com/xeipuuv/gojsonschema"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type Form struct {
	typ   typesystem.Type
	valid bool

	Schema *jsonschema.Schema
	Value  any

	validators []FormValidator
}

type FormOption func(f *Form)

type FormValidator func(f *Form) (*gojsonschema.Result, error)

func WithSchema(schema *jsonschema.Schema) FormOption {
	return func(f *Form) {
		f.SetSchema(schema)
	}
}

func WithValidator(validator FormValidator) FormOption {
	return func(f *Form) {
		f.validators = append(f.validators, validator)
	}
}

func WithSchemaFor[T any]() FormOption {
	return WithSchemaFromType(typesystem.GetType[T]())

}

func WithSchemaFromType(typ typesystem.Type) FormOption {
	return func(f *Form) {
		f.typ = typ
		f.Schema = typesystem.Universe().CompileFlatBundleFor(typ.JsonSchema())
	}
}

func NewForm(opts ...FormOption) *Form {
	f := &Form{
		Schema: &jsonschema.Schema{},
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (f *Form) GetData() any                  { return f.Value }
func (f *Form) GetSchema() *jsonschema.Schema { return f.Schema }
func (f *Form) SetSchema(schema *jsonschema.Schema) {
	f.Schema = schema

	f.Invalidate()
}

func (f *Form) MustSetField(field string, value any) {
	if err := f.SetField(field, value); err != nil {
		panic(err)
	}
}

func (f *Form) MustGetField(field string) any {
	v, _, err := f.GetField(field)

	if err != nil {
		panic(err)
	}

	return v
}

func (f *Form) GetField(field string) (any, bool, error) {
	ptr, err := gojsonpointer.NewJsonPointer(field)

	if err != nil {
		return nil, false, err
	}

	v, _, err := ptr.Get(f.Value)

	if err != nil {
		return nil, false, err
	}

	return v, true, nil
}

func (f *Form) SetField(field string, value any) error {
	ptr, err := NewJsonPointer(field)

	if err != nil {
		return err
	}

	rootValue := reflect.ValueOf(&f.Value)
	currentValue := rootValue
	currentSchema := f.Schema

	createDefaultValue := func(schema *jsonschema.Schema) any {
		switch schema.Type {
		case "object":
			return make(map[string]any)
		case "array":
			return make([]any, 0)
		case "string":
			return ""
		case "number":
			return 0.0
		case "integer":
			return 0
		case "boolean":
			return false
		default:
			return nil
		}
	}

	replaceCurrentValue := func(v reflect.Value) {
		if !v.IsValid() {
			rootValue.Elem().Set(reflect.Zero(rootValue.Type()))
			return
		}

		rootValue.Elem().Set(v)
	}

	for _, token := range ptr.Elements() {
		var nextValue reflect.Value

		nextSchema := f.getSchemaForPath(currentSchema, token)

		if nextSchema == nil {
			return errors.New("invalid field")
		}

		switch currentSchema.Type {
		case "object":
			k := reflect.ValueOf(token)

			if !currentValue.IsValid() || currentValue.IsNil() {
				currentValue = reflect.ValueOf(createDefaultValue(currentSchema))
				replaceCurrentValue(currentValue)
			}

			for currentValue.Kind() == reflect.Ptr || currentValue.Kind() == reflect.Interface {
				currentValue = currentValue.Elem()
			}

			if !currentValue.IsValid() || currentValue.IsNil() {
				currentValue = reflect.ValueOf(createDefaultValue(currentSchema))
				replaceCurrentValue(currentValue)
			}

			this := currentValue
			nextValue = currentValue.MapIndex(k)

			if !nextValue.IsValid() {
				nextValue = reflect.ValueOf(createDefaultValue(nextSchema))
				this.SetMapIndex(k, nextValue)
			} else {
				nextValue = nextValue.Elem()
			}

			replaceCurrentValue = func(v reflect.Value) {
				this.SetMapIndex(k, v)
			}

		case "array":
			index, err := strconv.Atoi(token)

			if err != nil {
				return err
			}

			if index < 0 {
				return errors.New("invalid field index (must be >= 0)")
			}

			currentLen := 0

			if currentValue.Kind() == reflect.Ptr || currentValue.Kind() == reflect.Interface {
				currentValue = currentValue.Elem()
			}

			if !currentValue.IsNil() {
				currentLen = currentValue.Len()
			}

			if index >= currentLen {
				currentValue = reflect.AppendSlice(currentValue, reflect.MakeSlice(anySlice, index-currentLen+1, index-currentLen+1))
				replaceCurrentValue(currentValue)
			}

			this := currentValue.Index(index)
			nextValue = this

			if !nextValue.IsValid() {
				nextValue = reflect.ValueOf(createDefaultValue(nextSchema))
				this.Index(index).Set(nextValue)
			} else {
				nextValue = nextValue.Elem()
			}

			replaceCurrentValue = func(v reflect.Value) {
				this.Index(index).Set(v)
			}

		default:
			return errors.New("invalid field")
		}

		currentSchema = nextSchema
		currentValue = nextValue
	}

	replaceCurrentValue(reflect.ValueOf(value))

	f.Invalidate()

	return nil
}

func (f *Form) MergeFields(from any) error {
	result, err := deepmerge.Merge(from, f.Value)

	if err != nil {
		return err
	}

	f.Value = result

	f.Invalidate()

	return nil
}

func (f *Form) MarshalFrom(v any) error {
	var mapped any

	data, err := json.Marshal(v)

	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &mapped)

	if err != nil {
		return err
	}

	if f.Value == nil {
		f.Value = mapped

		return nil
	}

	return f.MergeFields(mapped)
}

func (f *Form) UnmarshalTo(v any) error {
	data, err := json.Marshal(f.Value)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (f *Form) UnmarshalWithPrototype(proto ipld.NodePrototype) (ipld.Node, error) {
	data, err := json.Marshal(f.Value)

	if err != nil {
		return nil, err
	}

	return ipld.DecodeUsingPrototype(data, dagjson.Decode, proto)
}

func (f *Form) Build() (any, error) {
	if f.typ != nil {
		n, err := f.UnmarshalWithPrototype(f.typ.IpldPrototype())

		if err != nil {
			return nil, err
		}

		return typesystem.Unwrap(n), nil
	}

	return f.Value, nil
}

func (f *Form) Invalidate() {
	f.valid = false
}

func (f *Form) Validate() (*gojsonschema.Result, error) {
	f.valid = false

	bundle := typesystem.Universe().CompileBundleFor(f.Schema)
	serializedSchema, err := bundle.MarshalJSON()

	if err != nil {
		return nil, err
	}

	schema := gojsonschema.NewStringLoader(string(serializedSchema))
	document := gojsonschema.NewGoLoader(f.Value)

	result, err := gojsonschema.Validate(schema, document)

	if err != nil {
		return nil, err
	}

	if !result.Valid() {
		return result, nil
	}

	for _, validator := range f.validators {
		result, err = validator(f)

		if err != nil {
			return nil, err
		}

		if !result.Valid() {
			return result, nil
		}
	}

	f.valid = true

	return result, nil
}

func (f *Form) Clear() {
	f.Value = make(map[string]any)
	f.valid = false
}

func (f *Form) SubForm(path string) (*Form, error) {
	ptr, err := NewJsonPointer(path)

	if err != nil {
		return nil, err
	}

	schema := f.getSchemaForPath(f.Schema, ptr.Elements()...)

	if schema == nil {
		return nil, errors.New("invalid path")
	}

	doc, ok, err := f.GetField(path)

	if err != nil {
		return nil, err
	}

	if !ok {
		doc = make(map[string]any)

		if err := f.SetField(path, doc); err != nil {
			return nil, err
		}
	}

	return &Form{
		Value:  doc.(map[string]any),
		Schema: schema,
	}, nil
}

func (f *Form) getSchemaForPath(root *jsonschema.Schema, path ...string) *jsonschema.Schema {
	s := root

	resolveSchema := func(schema *jsonschema.Schema) *jsonschema.Schema {
		if schema.Ref == "" {
			return schema
		}

		parsed, err := gojsonreference.NewJsonReference(schema.Ref)

		if err != nil {
			return nil
		}

		name := parsed.GetPointer().String()
		name = strings.TrimPrefix(name, "/$defs/")
		name = strings.TrimPrefix(name, "/definitions/")

		return typesystem.Universe().LookupByName(name).JsonSchema()
	}

	for _, token := range path {
		if s == nil {
			return nil
		}

		switch s.Type {
		case "object":
			v, ok := s.Properties.Get(token)

			if !ok {
				return nil
			}

			vm, ok := v.(*jsonschema.Schema)

			if !ok {
				return nil
			}

			if vm == nil {
				return nil
			}

			s = resolveSchema(vm)
		case "array":
			s = resolveSchema(s.Items)
		default:
			return nil
		}
	}

	return s
}

func (f *Form) UnmarshalJSON(raw []byte) error {
	return json.Unmarshal(raw, &f.Value)
}

func (f *Form) MarshalJSON() ([]byte, error) {
	return json.Marshal(f.Value)
}

func FillForm[T any](ctx context.Context, history PromptMessageSource, result *T, opts ...FormOption) error {
	f := NewForm(WithSchemaFor[T]())

	for _, opt := range opts {
		opt(f)
	}

	if err := f.MarshalFrom(result); err != nil {
		return err
	}

	ft := NewFormTask(f, history)

	if _, err := ft.RunToCompletion(ctx); err != nil {
		return err
	}

	n, err := f.Build()

	if err != nil {
		return err
	}

	*result = n.(T)

	return nil
}
