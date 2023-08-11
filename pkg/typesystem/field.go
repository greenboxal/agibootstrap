package typesystem

import "reflect"

type fieldBase struct {
	declaringType StructType
	name          string
	typ           Type
	virtual       bool
}

func (f *fieldBase) Name() string              { return f.name }
func (f *fieldBase) Type() Type                { return f.typ }
func (f *fieldBase) DeclaringType() StructType { return f.declaringType }
func (f *fieldBase) IsVirtual() bool           { return false }

type reflectedField struct {
	fieldBase

	runtimeField reflect.StructField
}

func (f *reflectedField) Resolve(receiver Value) Value {
	v := reflect.Indirect(receiver.Value()).FieldByIndex(f.runtimeField.Index)

	return ValueFrom(v).As(f.typ)
}
