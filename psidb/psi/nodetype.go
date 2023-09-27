package psi

import (
	"context"
	"io"
	"reflect"

	"github.com/ipld/go-ipld-prime"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type NodeClass string

const (
	NodeClassInvalid  NodeClass = ""
	NodeClassGeneric  NodeClass = "generic"
	NodeClassCode     NodeClass = "code"
	NodeClassDocument NodeClass = "document"
)

type NodeTypeDefinition struct {
	Name  string    `json:"name"`
	Class NodeClass `json:"class"`

	IsRuntimeOnly bool `json:"is_runtime_only"`
	IsStub        bool `json:"is_stub"`
}

type NodeTypeOption func(*nodeType)

func WithStubNodeType() NodeTypeOption {
	return func(nt *nodeType) {
		nt.def.IsStub = true
	}
}

func WithRuntimeOnly() NodeTypeOption {
	return func(nt *nodeType) {
		nt.def.IsRuntimeOnly = true
	}
}

func WithInterface(iface NodeInterface, impl VTableDefinition) NodeTypeOption {
	return func(nt *nodeType) {
		nt.vtables[iface.Name()] = BindInterface(iface, impl)
	}
}

func WithInterfaceFromNode(iface NodeInterface) NodeTypeOption {
	return func(nt *nodeType) {
		nt.vtables[iface.Name()] = BindInterfaceFromNode(iface, nt.typ)
	}
}

type NodeType interface {
	Name() string
	Type() typesystem.Type
	TypeName() typesystem.TypeName
	RuntimeType() reflect.Type
	Definition() NodeTypeDefinition

	CreateInstance() Node
	InitializeNode(n Node)

	OnAfterNodeLoaded(ctx context.Context, n Node) error
	OnBeforeNodeSaved(ctx context.Context, n Node) error

	EncodeNode(w io.Writer, encoder ipld.Encoder, n Node) error
	DecodeNode(r io.Reader, decoder ipld.Decoder) (Node, error)

	String() string

	Interfaces() []*VTable
	Interface(name string) *VTable
}

type TypedNodeType[T Node] interface {
	NodeType
}

type typedNode[T Node] struct {
	NodeType
}

type nodeType struct {
	name    typesystem.TypeName
	typ     typesystem.Type
	def     NodeTypeDefinition
	vtables map[string]*VTable
}

func (nt *nodeType) OnBeforeNodeSaved(ctx context.Context, n Node) error {
	return nil
}

func (nt *nodeType) OnAfterNodeLoaded(ctx context.Context, n Node) error {
	nv := reflect.ValueOf(n)

	for nv.Kind() == reflect.Ptr {
		nv = reflect.Indirect(nv)
	}

	t := nv.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag := f.Tag.Get("psi-edge")

		if tag == "" {
			continue
		}

		name, err := ParsePathElement(tag)

		if err != nil {
			return err
		}

		fieldTypePtr := f.Type

		if fieldTypePtr.Kind() != reflect.Ptr {
			fieldTypePtr = reflect.PtrTo(fieldTypePtr)
		}

		if fieldTypePtr.Implements(mutableNodeReferenceType) {
			var ref MutableNodeReference

			v := nv.Field(i)

			if v.Type().Kind() == reflect.Ptr {
				if v.IsNil() {
					v.Set(reflect.New(v.Type().Elem()))
				}

				ref = v.Interface().(MutableNodeReference)
			} else {
				ref = v.Addr().Interface().(MutableNodeReference)
			}

			p := n.CanonicalPath().Child(name)

			if err := ref.SetPathReference(ctx, p); err != nil {
				return err
			}
		} else {
			return errors.New("field does not implement MutableNodeReference")
		}
	}

	return nil
}

type MutableNodeReference interface {
	SetNodeReference(ctx context.Context, n Node) error
	SetPathReference(ctx context.Context, p Path) error
}

var mutableNodeReferenceType = reflect.TypeOf((*MutableNodeReference)(nil)).Elem()

func (nt *nodeType) Name() string                   { return nt.name.MangledName() }
func (nt *nodeType) Type() typesystem.Type          { return nt.typ }
func (nt *nodeType) TypeName() typesystem.TypeName  { return nt.name }
func (nt *nodeType) RuntimeType() reflect.Type      { return nt.typ.RuntimeType() }
func (nt *nodeType) Definition() NodeTypeDefinition { return nt.def }
func (nt *nodeType) Interfaces() []*VTable          { return maps.Values(nt.vtables) }
func (nt *nodeType) Interface(name string) *VTable  { return nt.vtables[name] }
func (nt *nodeType) String() string                 { return nt.Name() }

func (nt *nodeType) CreateInstance() Node {
	return reflect.New(nt.typ.RuntimeType()).Interface().(Node)
}

func (nt *nodeType) InitializeNode(n Node) {
	if init, ok := n.(baseNodeInitializer); ok {
		init.Init(n)
	} else if init, ok := n.(baseNodeInitializerWithOptions); ok {
		init.Init(n, WithNodeType(nt))
	}
}

func (nt *nodeType) EncodeNode(w io.Writer, encoder ipld.Encoder, n Node) error {
	node := typesystem.Wrap(n)

	return ipld.EncodeStreaming(w, node, encoder)
}

func (nt *nodeType) DecodeNode(r io.Reader, decoder ipld.Decoder) (Node, error) {
	proto := nt.Type().IpldPrototype()
	node, err := ipld.DecodeStreamingUsingPrototype(r, decoder, proto)

	if err != nil {
		return nil, err
	}

	return typesystem.Unwrap(node).(Node), nil
}

type baseNodeInitializer interface {
	Init(n Node)
}

type baseNodeInitializerWithOptions interface {
	Init(n Node, options ...NodeInitOption)
}

type baseNodeOnInitialize interface {
	NodeType

	OnInitNode(n Node)
}
