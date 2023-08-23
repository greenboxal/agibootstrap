package typesystem

import (
	"encoding"
	"errors"
	"reflect"
	"strconv"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
)

type reprBuilder struct{ nodeBuilder }

type nodeBuilder struct {
	v Value
}

func newNodeBuilder(v Value) *nodeBuilder {
	if v.typ == nil {
		panic("no node typ set")
	}

	return &nodeBuilder{v: v}
}

func (bb *nodeBuilder) BeginMap(sizeHint int64) (datamodel.MapAssembler, error) {
	if sizeHint < 0 {
		sizeHint = 0
	}

	switch bb.v.Type().PrimitiveKind() {
	case PrimitiveKindMap:
		mt := bb.v.Type().Map()

		if !bb.v.Value().IsValid() {
			bb.v = MakeMap(mt.Key(), mt.Value(), int(sizeHint))
		}

		if bb.v.v.IsNil() {
			bb.v.v.Set(MakeMap(mt.Key(), mt.Value(), int(sizeHint)).v)
		}

		return &mapAssembler{bb: bb}, nil

	case PrimitiveKindStruct:
		if !bb.v.Value().IsValid() {
			bb.v = New(bb.v.Type())
		}

		if bb.v.v.Kind() == reflect.Ptr && bb.v.v.IsNil() {
			bb.v.v.Set(New(bb.v.Type()).v.Addr())
		}

		return &structAssembler{bb: bb}, nil

	case PrimitiveKindInterface:
		return &ifaceBuilder{expected: bb.v.Type(), bb: bb}, nil
	}

	return nil, errors.New("cannot begin map on non-map type")
}

func (bb *nodeBuilder) BeginList(sizeHint int64) (datamodel.ListAssembler, error) {
	if sizeHint < 0 {
		sizeHint = 0
	}

	if !bb.v.v.IsValid() {
		bb.v = MakeList(bb.v.Type().List().Elem(), 0, int(sizeHint))
		bb.v.v = bb.v.v.Addr()
	}

	return &listAssembler{bb: bb}, nil
}

func (bb *nodeBuilder) AssignNull() error {
	if bb.v.v.Kind() == reflect.Ptr {
		bb.v.v.Set(reflect.Zero(bb.v.v.Type()))
	} else {
		bb.v.Indirect().Set(reflect.Zero(bb.v.Type().RuntimeType()))
	}

	return nil
}

func (bb *nodeBuilder) AssignBool(b bool) error {
	if bb.v.v.Kind() == reflect.Ptr && bb.v.v.IsNil() {
		bb.v.v.Set(reflect.New(bb.v.v.Type().Elem()))
	}

	bb.v.Indirect().SetBool(b)

	return nil
}

func (bb *nodeBuilder) AssignInt(i int64) error {
	if bb.v.v.Kind() == reflect.Ptr && bb.v.v.IsNil() {
		bb.v.v.Set(reflect.New(bb.v.v.Type().Elem()))
	}

	if bb.v.Type().PrimitiveKind() == PrimitiveKindFloat {
		bb.v.Indirect().SetFloat(float64(float32(i)))
	} else if bb.v.Type().PrimitiveKind() == PrimitiveKindUnsignedInt {
		bb.v.Indirect().SetUint(uint64(i))
	} else {
		bb.v.Indirect().SetInt(i)
	}

	return nil
}

func (bb *nodeBuilder) AssignFloat(f float64) error {
	if bb.v.v.Kind() == reflect.Ptr && bb.v.v.IsNil() {
		bb.v.v.Set(reflect.New(bb.v.v.Type().Elem()))
	}

	bb.v.Indirect().SetFloat(f)

	return nil
}

func (bb *nodeBuilder) AssignString(s string) error {
	if bb.v.v.Kind() == reflect.Ptr && bb.v.v.IsNil() {
		bb.v.v.Set(reflect.New(bb.v.v.Type().Elem()))
	}

	v := bb.v.Indirect()

	switch bb.v.typ.PrimitiveKind() {
	case PrimitiveKindString:
		v.SetString(s)

	case PrimitiveKindInt:
		i, err := strconv.ParseInt(s, 10, 64)

		if err != nil {
			return err
		}

		v.SetInt(i)

	case PrimitiveKindUnsignedInt:
		i, err := strconv.ParseUint(s, 10, 64)

		if err != nil {
			return err
		}

		v.SetUint(i)

	case PrimitiveKindFloat:
		f, err := strconv.ParseFloat(s, 64)

		if err != nil {
			return err
		}

		v.SetFloat(f)

	default:
		if v.Type().Implements(textUnmarshalerType) {
			if v.Kind() == reflect.Ptr && v.IsNil() {
				v.Set(reflect.New(v.Type()))
			}

			u := v.Interface().(encoding.TextUnmarshaler)

			return u.UnmarshalText([]byte(s))
		} else if reflect.PtrTo(v.Type()).Implements(textUnmarshalerType) {
			u := v.Addr().Interface().(encoding.TextUnmarshaler)

			return u.UnmarshalText([]byte(s))
		} else if u, ok := TryCast[encoding.TextUnmarshaler](v); ok {
			return u.UnmarshalText([]byte(s))
		}

		return errors.New("cannot assign string to non-string type")
	}

	return nil
}

func (bb *nodeBuilder) AssignBytes(bytes []byte) error {
	if bb.v.v.Kind() == reflect.Ptr && bb.v.v.IsNil() {
		bb.v.v.Set(reflect.New(bb.v.v.Type().Elem()))
	}

	v := bb.v.Indirect()

	if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
		v.SetBytes(bytes)

		return nil
	} else if u, ok := TryCast[encoding.BinaryUnmarshaler](v); ok {
		if len(bytes) == 0 {
			return nil
		}

		return u.UnmarshalBinary(bytes)
	} else {
		return errors.New("cannot assign bytes to non-bytes type")
	}
}

func (bb *nodeBuilder) AssignLink(link datamodel.Link) error {
	if bb.v.v.Kind() == reflect.Ptr && bb.v.v.IsNil() {
		bb.v.v.Set(reflect.New(bb.v.v.Type().Elem()))
	}

	bb.v.Indirect().Set(reflect.ValueOf(link))

	return nil
}

func (bb *nodeBuilder) AssignNode(node datamodel.Node) error {
	if bb.v.v.Kind() == reflect.Ptr && bb.v.v.IsNil() {
		bb.v.v.Set(reflect.New(bb.v.v.Type().Elem()))
	}

	vn := reflect.ValueOf(node)
	vt := bb.v.Type()

	switch node.Kind() {
	case datamodel.Kind_Null:
		return bb.AssignNull()
	case datamodel.Kind_Bool:
		v, err := node.AsBool()
		if err != nil {
			return err
		}
		return bb.AssignBool(v)
	case datamodel.Kind_Int:
		v, err := node.AsInt()
		if err != nil {
			return err
		}
		return bb.AssignInt(v)
	case datamodel.Kind_String:
		v, err := node.AsString()
		if err != nil {
			return err
		}
		return bb.AssignString(v)
	case datamodel.Kind_Float:
		v, err := node.AsFloat()
		if err != nil {
			return err
		}
		return bb.AssignFloat(v)
	case datamodel.Kind_Bytes:
		v, err := node.AsBytes()
		if err != nil {
			return err
		}
		return bb.AssignBytes(v)
	case datamodel.Kind_Link:
		v, err := node.AsLink()
		if err != nil {
			return err
		}
		return bb.AssignLink(v)

	case datamodel.Kind_List:
		it := node.ListIterator()
		count := node.Length()

		lb, err := bb.BeginList(count)

		if err != nil {
			return err
		}

		for !it.Done() {
			_, n, err := it.Next()

			if err != nil {
				return err
			}

			if err := lb.AssembleValue().AssignNode(n); err != nil {
				return err
			}
		}

		return lb.Finish()

	case datamodel.Kind_Map:
		it := node.MapIterator()
		count := node.Length()

		mb, err := bb.BeginMap(count)

		if err != nil {
			return err
		}

		for !it.Done() {
			k, v, err := it.Next()

			if err != nil {
				return err
			}

			if err := mb.AssembleKey().AssignNode(k); err != nil {
				return err
			}

			if err := mb.AssembleValue().AssignNode(v); err != nil {
				return err
			}
		}

		return mb.Finish()

	default:
		if vn.Type().AssignableTo(vt.RuntimeType()) {
			bb.v.Indirect().Set(vn)
		}

		return errors.New("cannot assign node")
	}

	return nil
}

func (bb *nodeBuilder) Prototype() datamodel.NodePrototype {
	return bb.v.Type().IpldPrototype()
}

func (bb *nodeBuilder) Build() datamodel.Node {
	return bb.v.AsNode()
}

func (bb *nodeBuilder) Reset() {
	bb.v = New(bb.v.Type())
}

type listAssembler struct {
	bb *nodeBuilder

	nextValue *Value
}

func (la *listAssembler) AssembleValue() datamodel.NodeAssembler {
	lt := la.bb.v.Type().(ListType)

	v := New(lt.Elem())
	l := la.bb.v.Value()

	i := l.Len()
	l = reflect.Append(l, v.Value())

	reflect.Indirect(la.bb.v.v).Set(l)

	v = ValueFrom(l.Index(i))

	return newNodeBuilder(v)
}

func (la *listAssembler) ValuePrototype(idx int64) datamodel.NodePrototype {
	return la.bb.v.Type().List().Elem().IpldPrototype()
}

func (la *listAssembler) Finish() error {
	la.next()

	return nil
}

func (la *listAssembler) next() {
	if la.nextValue != nil {
		la.bb.v.Value().Set(reflect.Append(la.bb.v.Value(), la.nextValue.Value()))

		la.nextValue = nil
	}

	v := New(la.bb.v.Type().(ListType).Elem())

	la.nextValue = &v
}

type mapAssembler struct {
	bb *nodeBuilder

	nextKey   *Value
	nextValue *Value
}

func (ma *mapAssembler) AssembleKey() datamodel.NodeAssembler {
	ma.next()

	return newNodeBuilder(*ma.nextKey)
}

func (ma *mapAssembler) AssembleValue() datamodel.NodeAssembler {
	return newNodeBuilder(*ma.nextValue)
}

func (ma *mapAssembler) AssembleEntry(k string) (datamodel.NodeAssembler, error) {
	if err := ma.AssembleKey().AssignString(k); err != nil {
		return nil, err
	}

	return ma.AssembleValue(), nil
}

func (ma *mapAssembler) KeyPrototype() datamodel.NodePrototype {
	return ma.bb.v.Type().Map().Key().IpldPrototype()
}

func (ma *mapAssembler) ValuePrototype(k string) datamodel.NodePrototype {
	return ma.bb.v.Type().Map().Value().IpldPrototype()
}

func (ma *mapAssembler) Finish() error {
	ma.next()

	return nil
}

func (ma *mapAssembler) next() {
	if ma.nextValue != nil {
		ma.bb.v.Value().SetMapIndex(ma.nextKey.Value(), ma.nextValue.Value())

		ma.nextKey = nil
		ma.nextValue = nil
	}

	k := New(ma.bb.v.Type().Map().Key())
	ma.nextKey = &k

	v := New(ma.bb.v.Type().Map().Value())
	ma.nextValue = &v
}

type structAssembler struct {
	bb  *nodeBuilder
	key *Value
}

func (sa *structAssembler) AssembleKey() datamodel.NodeAssembler {
	if sa.key == nil {
		k := New(TypeOf(""))
		sa.key = &k
	}

	return newNodeBuilder(*sa.key)
}

func (sa *structAssembler) AssembleValue() datamodel.NodeAssembler {
	if sa.key == nil {
		panic("AssembleValue called before AssembleKey")
	}

	name := sa.key.Value().String()

	st := sa.bb.v.typ.Struct()
	fld := st.Field(name)

	if fld == nil {
		return newNodeBuilder(New(TypeFrom(reflect.TypeOf((*any)(nil)).Elem())))
	}

	v := fld.Resolve(sa.bb.v)

	return newNodeBuilder(v)
}

func (sa *structAssembler) AssembleEntry(k string) (datamodel.NodeAssembler, error) {
	if err := sa.AssembleKey().AssignString(k); err != nil {
		return nil, err
	}

	return sa.AssembleValue(), nil
}

func (sa *structAssembler) KeyPrototype() datamodel.NodePrototype {
	return basicnode.Prototype.String
}

func (sa *structAssembler) ValuePrototype(k string) datamodel.NodePrototype {
	fld := sa.bb.v.typ.Struct().Field(k)

	return fld.Type().IpldPrototype()
}

func (sa *structAssembler) Finish() error {
	return nil
}
