package psi

import "reflect"

type NodeClass string

const (
	NodeClassInvalid  NodeClass = ""
	NodeClassGeneric  NodeClass = "generic"
	NodeClassCode     NodeClass = "code"
	NodeClassDocument NodeClass = "document"
)

var nodeTypeRegistry = map[string]NodeType{}

type NodeTypeOption func(*nodeType)

func RegisterNodeType[T Node](name string, options ...NodeTypeOption) TypedNodeType[T] {
	nt := &nodeType{
		name:  name,
		class: NodeClassGeneric,
		typ:   reflect.TypeOf((*T)(nil)).Elem(),
	}

	for _, opt := range options {
		opt(nt)
	}

	nodeTypeRegistry[name] = nt

	return nt
}

func WithNodeClass(class NodeClass) NodeTypeOption {
	return func(nt *nodeType) {
		nt.class = class
	}
}

type NodeType interface {
	Name() string
	Class() NodeClass
	RuntimeType() reflect.Type
}

type TypedNodeType[T Node] interface {
	NodeType
}

type typedNode[T Node] struct {
	nodeType
}

type nodeType struct {
	name  string
	class NodeClass
	typ   reflect.Type
}

func (n *nodeType) Name() string              { return n.name }
func (n *nodeType) Class() NodeClass          { return n.class }
func (n *nodeType) RuntimeType() reflect.Type { return n.typ }
