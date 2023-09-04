package typesystem

import "reflect"

type fieldBase struct {
	declaringType StructType
	name          string
	typ           Type

	virtual  bool
	optional bool
	nullable bool
}

func (f *fieldBase) Name() string              { return f.name }
func (f *fieldBase) Type() Type                { return f.typ }
func (f *fieldBase) DeclaringType() StructType { return f.declaringType }
func (f *fieldBase) IsVirtual() bool           { return f.virtual }
func (f *fieldBase) IsNullable() bool          { return f.nullable }
func (f *fieldBase) IsOptional() bool          { return f.optional }

type reflectedField struct {
	fieldBase

	runtimeField reflect.StructField
}

func (f *reflectedField) Resolve(receiver Value) Value {
	v := reflect.Indirect(receiver.Value()).FieldByIndex(f.runtimeField.Index)

	return ValueFrom(v).As(f.typ)
}
