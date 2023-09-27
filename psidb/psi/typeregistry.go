package psi

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/greenboxal/aip/aip-sdk/pkg/utils"
	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

var globalTypeRegistry = NewTypeRegistry()

func GlobalTypeRegistry() TypeRegistry {
	return globalTypeRegistry
}

type TypeRegistry interface {
	NodeTypes() []NodeType
	EdgeTypes() []EdgeType

	RegisterNodeType(nt NodeType)
	RegisterEdgeKind(kind EdgeKind, options ...EdgeTypeOption)
	ReflectNodeType(typ typesystem.Type) NodeType

	LookupEdgeType(kind EdgeKind) EdgeType

	NodeTypeByName(ctx context.Context, name string) NodeType
}

type typeRegistry struct {
	nodeTypeRegistryMutex sync.RWMutex
	nodeTypeRegistry      map[typesystem.Type]NodeType
	nodeTypeByName        map[string]NodeType

	edgeTypeRegistryMutex sync.RWMutex
	edgeTypeByName        map[EdgeKind]EdgeType
}

func NewTypeRegistry() TypeRegistry {
	return &typeRegistry{
		nodeTypeRegistry: map[typesystem.Type]NodeType{},
		nodeTypeByName:   map[string]NodeType{},

		edgeTypeByName: map[EdgeKind]EdgeType{},
	}
}

func (tr *typeRegistry) NodeTypes() []NodeType {
	tr.nodeTypeRegistryMutex.RLock()
	defer tr.nodeTypeRegistryMutex.RUnlock()

	return maps.Values(tr.nodeTypeRegistry)
}

func (tr *typeRegistry) EdgeTypes() []EdgeType {
	tr.edgeTypeRegistryMutex.RLock()
	defer tr.edgeTypeRegistryMutex.RUnlock()

	return maps.Values(tr.edgeTypeByName)
}

func (tr *typeRegistry) RegisterNodeType(nt NodeType) {
	tr.nodeTypeRegistryMutex.Lock()
	defer tr.nodeTypeRegistryMutex.Unlock()

	if _, ok := tr.nodeTypeRegistry[nt.Type()]; ok {
		panic("node type already defined")
	}

	tr.nodeTypeRegistry[nt.Type()] = nt
	tr.nodeTypeByName[nt.Name()] = nt
}

func (tr *typeRegistry) NodeTypeByName(ctx context.Context, name string) NodeType {
	tr.nodeTypeRegistryMutex.RLock()
	defer tr.nodeTypeRegistryMutex.RUnlock()

	return tr.nodeTypeByName[name]
}

func (tr *typeRegistry) ReflectNodeType(typ typesystem.Type) NodeType {
	tr.nodeTypeRegistryMutex.Lock()
	defer tr.nodeTypeRegistryMutex.Unlock()

	if _, ok := tr.nodeTypeRegistry[typ]; ok {
		return tr.nodeTypeRegistry[typ]
	}

	nt := reflectNodeType(typ)

	tr.nodeTypeRegistry[typ] = nt
	tr.nodeTypeByName[nt.Name()] = nt

	return nt
}

func (tr *typeRegistry) RegisterEdgeType(et EdgeType) {
	tr.edgeTypeRegistryMutex.Lock()
	defer tr.edgeTypeRegistryMutex.Unlock()

	if _, ok := tr.edgeTypeByName[et.Kind()]; ok {
		panic("edge type already defined")
	}

	tr.edgeTypeByName[et.Kind()] = et
}

func (tr *typeRegistry) RegisterEdgeKind(kind EdgeKind, options ...EdgeTypeOption) {
	et := &edgeType{}
	et.kind = kind
	et.name = string(kind)

	for _, opt := range options {
		opt(et)
	}

	tr.RegisterEdgeType(et)
}

func (tr *typeRegistry) LookupEdgeType(kind EdgeKind) EdgeType {
	tr.edgeTypeRegistryMutex.RLock()
	defer tr.edgeTypeRegistryMutex.RUnlock()

	return tr.edgeTypeByName[kind]
}

func DefineNodeType[T Node](options ...NodeTypeOption) TypedNodeType[T] {
	rt := reflect.TypeOf((*T)(nil)).Elem()
	typ := typesystem.TypeFrom(rt)
	nt := reflectNodeType(typ, options...)
	tnt := typedNode[T]{NodeType: nt}

	globalTypeRegistry.RegisterNodeType(tnt)

	return tnt
}

func LookupEdgeType(kind EdgeKind) EdgeType {
	return globalTypeRegistry.LookupEdgeType(kind)
}

func ReflectNodeType(typ typesystem.Type) NodeType {
	return globalTypeRegistry.ReflectNodeType(typ)
}

func DefineEdgeType[T Node](kind EdgeKind, options ...EdgeTypeOption) TypedEdgeKind[T] {
	globalTypeRegistry.RegisterEdgeKind(kind, options...)

	return TypedEdgeKind[T](kind)
}

func formatTypeName(typName typesystem.TypeName) string {
	pkg := typName.Package
	name := typName.Name
	args := ""

	if len(typName.InParameters) > 0 {
		args = "["
		for i, param := range typName.InParameters {
			if i > 0 {
				args += ", "
			}

			args += formatTypeName(param)
		}
		args += "]"
	}

	return fmt.Sprintf("%s.%s%s", pkg, name, args)
}

func reflectNodeType(typ typesystem.Type, options ...NodeTypeOption) *nodeType {
	nt := &nodeType{
		typ:  typ,
		name: typ.Name(),

		vtables: map[string]*VTable{},
	}

	nt.def.Class = NodeClassGeneric
	nt.def.Name = formatTypeName(typ.Name())
	nt.name = typesystem.AsTypeName(utils.ParseTypeName(nt.def.Name))

	//nt.vtables["Node"] = BindInterfaceFromNode(INodeInterface, typ)

	for _, opt := range options {
		opt(nt)
	}

	return nt
}
