package psi

import (
	"reflect"
	"sync"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
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
}

var nodeTypeRegistryMutex sync.Mutex
var nodeTypeRegistry = map[typesystem.Type]NodeType{}
var nodeTypeByName = map[string]NodeType{}

type NodeTypeOption func(*nodeType)

func DefineNodeType[T Node](options ...NodeTypeOption) TypedNodeType[T] {
	nodeTypeRegistryMutex.Lock()
	defer nodeTypeRegistryMutex.Unlock()

	rt := reflect.TypeOf((*T)(nil)).Elem()
	typ := typesystem.TypeFrom(rt)

	if _, ok := nodeTypeRegistry[typ]; ok {
		panic("node type already defined")
	}

	nt := reflectNodeType(typ, options...)
	tnt := typedNode[T]{NodeType: nt}

	nodeTypeRegistry[typ] = tnt
	nodeTypeByName[tnt.Name()] = tnt

	return tnt
}

func NodeTypeByName(name string) NodeType {
	return nodeTypeByName[name]
}

func ReflectNodeType(typ typesystem.Type) NodeType {
	nodeTypeRegistryMutex.Lock()
	defer nodeTypeRegistryMutex.Unlock()

	if _, ok := nodeTypeRegistry[typ]; ok {
		return nodeTypeRegistry[typ]
	}

	nt := reflectNodeType(typ)

	nodeTypeRegistry[typ] = nt
	nodeTypeByName[nt.Name()] = nt

	return nt
}

func reflectNodeType(typ typesystem.Type, options ...NodeTypeOption) *nodeType {
	nt := &nodeType{
		typ: typ,
	}

	nt.def.Name = nt.typ.Name().String()
	nt.def.Class = NodeClassGeneric

	for _, opt := range options {
		opt(nt)
	}

	return nt
}

func WithTypeName(name string) NodeTypeOption {
	return func(nt *nodeType) {
		nt.def.Name = name
	}
}

func WithNodeClass(class NodeClass) NodeTypeOption {
	return func(nt *nodeType) {
		nt.def.Class = class
	}
}

func WithRuntimeOnly() NodeTypeOption {
	return func(nt *nodeType) {
		nt.def.IsRuntimeOnly = true
	}
}

type NodeType interface {
	Name() string
	Type() typesystem.Type
	RuntimeType() reflect.Type
	Definition() NodeTypeDefinition
}

type TypedNodeType[T Node] interface {
	NodeType
}

type typedNode[T Node] struct {
	NodeType
}

type nodeType struct {
	typ typesystem.Type
	def NodeTypeDefinition
}

func (n *nodeType) Name() string                   { return n.def.Name }
func (n *nodeType) Type() typesystem.Type          { return n.typ }
func (n *nodeType) RuntimeType() reflect.Type      { return n.typ.RuntimeType() }
func (n *nodeType) Definition() NodeTypeDefinition { return n.def }
