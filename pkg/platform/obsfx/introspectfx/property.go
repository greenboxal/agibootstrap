package introspectfx

import (
	"reflect"

	"github.com/greenboxal/agibootstrap/pkg/platform/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/platform/obsfx/reflectfx"
)

type Property interface {
	Name() string
	Type() Type

	IsList() bool
	IsMap() bool
	IsSet() bool

	IsObservable() bool
	AsObservable(receiver Value) obsfx.Observable

	GetRawValue(receiver Value) reflect.Value
	GetValue(receiver Value) Value
}

type propertyBase struct {
	name        string
	runtimeType reflect.Type
	typ         Type
}

func (r *propertyBase) Name() string {
	return r.name
}

func (r *propertyBase) Type() Type {
	if r.typ == nil {
		r.typ = TypeFor(r.runtimeType)
	}

	return r.typ
}

func (r *propertyBase) IsList() bool {
	typ := r.runtimeType

	if typ.Kind() == reflect.Struct {
		typ = reflect.PointerTo(typ)
	}

	return reflectfx.IsObservableList(typ)
}

func (r *propertyBase) IsMap() bool {
	typ := r.runtimeType

	if typ.Kind() == reflect.Struct {
		typ = reflect.PointerTo(typ)
	}

	return reflectfx.IsObservableMap(typ)
}

func (r *propertyBase) IsSet() bool {
	typ := r.runtimeType

	if typ.Kind() == reflect.Struct {
		typ = reflect.PointerTo(typ)
	}

	return reflectfx.IsObservable(typ)
}

func (r *propertyBase) IsObservable() bool {
	typ := r.runtimeType

	if typ.Kind() == reflect.Struct {
		typ = reflect.PointerTo(typ)
	}

	return reflectfx.IsObservable(typ)
}

type methodProperty struct {
	propertyBase

	method reflect.Method
}

func (f *methodProperty) AsObservable(receiver Value) obsfx.Observable {
	return nil
}

func (f *methodProperty) GetRawValue(receiver Value) reflect.Value {
	v := receiver.Go()

	if v.IsNil() {
		return reflect.ValueOf(nil)
	}

	v = adjustPointers(f.method.Type.In(0), v)

	return f.method.Func.Call([]reflect.Value{v})[0]
}

func (f *methodProperty) GetValue(receiver Value) Value {
	return nestedValue{
		receiver: receiver,
		prop:     f,
	}
}

type funcProperty struct {
	propertyBase

	fn reflect.Value
}

func (f *funcProperty) AsObservable(receiver Value) obsfx.Observable {
	return nil
}

func (f *funcProperty) GetRawValue(receiver Value) reflect.Value {
	v := receiver.Go()

	if v.IsNil() {
		return reflect.ValueOf(nil)
	}

	return f.fn.Call([]reflect.Value{v})[0]
}

func (f *funcProperty) GetValue(receiver Value) Value {
	return nestedValue{
		receiver: receiver,
		prop:     f,
	}
}

type fieldProperty struct {
	propertyBase

	field reflect.StructField
}

func newReflectedFieldProperty(name string, field reflect.StructField) Property {
	return &fieldProperty{
		propertyBase: propertyBase{
			name:        name,
			runtimeType: field.Type,
		},

		field: field,
	}
}

func (r *fieldProperty) AsObservable(receiver Value) obsfx.Observable {
	if r.IsObservable() {
		v := r.GetRawValue(receiver)
		obs := v.Interface().(obsfx.Observable)

		return obs
	} else {
		return obsfx.BindExpression[any](func() any {
			v := r.GetRawValue(receiver)

			return v.Interface()
		})
	}
}

func (r *fieldProperty) GetRawValue(receiver Value) reflect.Value {
	self := receiver.Go()

	for self.Kind() == reflect.Ptr {
		if !self.IsValid() && !(self.CanAddr() && self.IsNil()) {
			return reflect.ValueOf(nil)
		}

		self = self.Elem()
	}

	if !self.IsValid() && !(self.CanAddr() && self.IsNil()) {
		return reflect.ValueOf(nil)
	}

	v := self.Field(r.field.Index[0])

	return v
}

func (r *fieldProperty) GetValue(receiver Value) Value {
	return nestedValue{
		receiver: receiver,
		prop:     r,
	}
}
