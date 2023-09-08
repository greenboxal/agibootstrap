package typesystem

import (
	"reflect"

	"github.com/ipld/go-ipld-prime"
)

type Value struct {
	typ Type
	v   reflect.Value
}

func (v Value) Value() reflect.Value      { return v.v }
func (v Value) Indirect() reflect.Value   { return reflect.Indirect(v.v) }
func (v Value) Type() Type                { return v.typ }
func (v Value) RuntimeType() reflect.Type { return v.typ.RuntimeType() }

func (v Value) As(typ Type) Value {
	v.typ = typ
	return v
}

func (v Value) AsNode() ipld.Node {
	if !v.v.IsValid() {

		return ipld.Null
	}

	if v.typ.RuntimeType() == reflect.TypeOf((*ipld.Node)(nil)).Elem() {
		if n, ok := TryCast[ipld.Node](v.Indirect()); ok {
			return n
		}
	}

	return newNode(v)
}

func (v Value) GetField(f Field) Value {
	return f.Resolve(v)
}
