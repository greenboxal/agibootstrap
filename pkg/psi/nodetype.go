package psi

import (
	"reflect"

	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
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
	RuntimeType() reflect.Type
	Definition() NodeTypeDefinition

	CreateInstance() Node
	InitializeNode(n Node)

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
	typ     typesystem.Type
	def     NodeTypeDefinition
	vtables map[string]*VTable
}

func (nt *nodeType) Name() string                   { return nt.def.Name }
func (nt *nodeType) Type() typesystem.Type          { return nt.typ }
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
