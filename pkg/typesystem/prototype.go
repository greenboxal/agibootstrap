package typesystem

import (
	"errors"
	"reflect"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/node/mixins"
	"github.com/ipld/go-ipld-prime/schema"
)

type valuePrototype struct {
	typ Type
}

func (v valuePrototype) NewBuilder() datamodel.NodeBuilder { return newNodeBuilder(New(v.typ)) }
func (v valuePrototype) Type() schema.Type                 { return v.typ.IpldType() }
func (v valuePrototype) Representation() datamodel.NodePrototype {
	return reprPrototype{typ: v.typ}
}

type ifacePrototype struct {
	typ Type
}

func (r ifacePrototype) NewBuilder() datamodel.NodeBuilder {
	return &ifaceBuilder{
		MapAssembler: mixins.MapAssembler{
			TypeName: r.typ.Name().NormalizedFullNameWithArguments(),
		},

		expected: r.typ,
	}
}

type reprPrototype struct {
	typ Type
}

func (r reprPrototype) NewBuilder() datamodel.NodeBuilder {
	return &reprBuilder{*newNodeBuilder(New(r.typ))}
}

type ifaceBuilder struct {
	mixins.MapAssembler

	bb *nodeBuilder

	expected Type
	actual   Type

	v Value
	k Value
	t Value
}

func (ib *ifaceBuilder) KeyPrototype() datamodel.NodePrototype {
	return basicnode.Prototype.String
}

func (ib *ifaceBuilder) ValuePrototype(k string) datamodel.NodePrototype {
	if k == "@type" {
		return basicnode.Prototype.String
	} else if k == "@value" {
		return ib.actual.IpldPrototype()
	}

	if f := ib.actual.Struct().Field(k); f != nil {
		return f.Type().IpldPrototype()
	}

	panic("invalid field")
}

func (ib *ifaceBuilder) AssembleKey() datamodel.NodeAssembler {
	if !ib.k.v.IsValid() {
		ib.k = New(TypeOf(""))
	}

	return newNodeBuilder(ib.k)
}

func (ib *ifaceBuilder) AssembleValue() datamodel.NodeAssembler {
	if !ib.k.v.IsValid() {
		panic("AssembleValue called before AssembleKey")
	}

	name := ib.k.v.String()

	if name == "" {
		panic("AssembleValue called before AssembleKey")
	}

	if name == "@type" {
		if !ib.t.v.IsValid() {
			ib.t = New(TypeOf(""))
		}

		return newNodeBuilder(ib.t)
	}

	if ib.actual == nil && ib.t.v.IsValid() {
		ib.actual = Universe().LookupByName(ib.t.v.String())
	}

	if !ib.v.v.IsValid() && ib.actual != nil {
		ib.v = New(ib.actual)
	}

	if name == "@value" {

	} else {
		st := ib.v.typ.Struct()
		fld := st.Field(name)

		if fld == nil {
			return newNodeBuilder(New(TypeFrom(reflect.TypeOf((*any)(nil)).Elem())))
		}

		v := fld.Resolve(ib.v)

		return newNodeBuilder(v)
	}

	panic("invalid field")
}

func (ib *ifaceBuilder) AssembleEntry(k string) (datamodel.NodeAssembler, error) {
	if err := ib.AssembleKey().AssignString(k); err != nil {
		return nil, err
	}

	return ib.AssembleValue(), nil
}

func (ib *ifaceBuilder) Finish() error {
	if ib.actual == nil && ib.t.v.IsValid() {
		ib.actual = Universe().LookupByName(ib.t.v.String())
	}

	if !ib.v.v.IsValid() && ib.actual != nil {
		ib.v = New(ib.actual)
	}

	if ib.v.v.Type().AssignableTo(ib.bb.v.v.Type()) {
		ib.bb.v.v.Set(ib.v.v)
	} else {
		ib.bb.v.v.Set(ib.v.v.Addr())
	}

	return nil
}

func (ib *ifaceBuilder) BeginMap(sizeHint int64) (datamodel.MapAssembler, error) {
	return ib, nil
}

func (ib *ifaceBuilder) AssignNode(node datamodel.Node) error {
	if node == ipld.Null {
		ib.v = Value{}
		return nil
	} else if n, ok := node.(valueNode); ok {
		ib.v = n.v
		return nil
	}

	return errors.New("invalid node")
}

func (ib *ifaceBuilder) Prototype() datamodel.NodePrototype {
	return ifacePrototype{typ: ib.expected}
}

func (ib *ifaceBuilder) Build() datamodel.Node {
	return ib.v.AsNode()
}

func (ib *ifaceBuilder) Reset() {
	ib.k = Value{}
	ib.v = Value{}
	ib.t = Value{}
	ib.actual = nil
}
