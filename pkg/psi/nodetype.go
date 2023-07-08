package psi

import (
	"reflect"

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

type NodeTypeOption func(*nodeType)

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

	CreateInstance() Node
	InitializeNode(n Node)
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

func (nt *nodeType) Name() string                   { return nt.def.Name }
func (nt *nodeType) Type() typesystem.Type          { return nt.typ }
func (nt *nodeType) RuntimeType() reflect.Type      { return nt.typ.RuntimeType() }
func (nt *nodeType) Definition() NodeTypeDefinition { return nt.def }

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
