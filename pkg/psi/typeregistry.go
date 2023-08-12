package psi

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

var globalTypeRegistry = NewTypeRegistry()

var packageTypeNameMap = map[string]string{
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs": "vfs",
	"github.com/greenboxal/agibootstrap/pkg/":             "agib.",
	"github.com/greenboxal/agibootstrap/psidb/db/":        "psidb.",
	"github.com/greenboxal/agibootstrap/psidb/services/":  "psidb.",
	"github.com/greenboxal/agibootstrap/psidb/modules/":   "",
}

func GlobalTypeRegistry() TypeRegistry {
	return globalTypeRegistry
}

type TypeRegistry interface {
	NodeTypes() []NodeType
	EdgeTypes() []EdgeType
	RegisterNodeType(nt NodeType)
	NodeTypeByName(name string) NodeType
	ReflectNodeType(typ typesystem.Type) NodeType
	LookupEdgeType(kind EdgeKind) EdgeType
	RegisterEdgeKind(kind EdgeKind, options ...EdgeTypeOption)
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

func (tr *typeRegistry) NodeTypeByName(name string) NodeType {
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

func NodeTypeByName(name string) NodeType {
	return globalTypeRegistry.NodeTypeByName(name)
}

func ReflectNodeType(typ typesystem.Type) NodeType {
	return globalTypeRegistry.ReflectNodeType(typ)
}

func LookupEdgeType(kind EdgeKind) EdgeType {
	return globalTypeRegistry.LookupEdgeType(kind)
}

func DefineEdgeType[T Node](kind EdgeKind, options ...EdgeTypeOption) TypedEdgeKind[T] {
	globalTypeRegistry.RegisterEdgeKind(kind, options...)

	return TypedEdgeKind[T](kind)
}

func rewritePackageName(pkg string) string {
	longest := ""

	for k := range packageTypeNameMap {
		if strings.HasPrefix(pkg, k) && len(k) > len(longest) {
			longest = k
		}
	}

	if longest == "" {
		return pkg
	}

	return packageTypeNameMap[longest] + strings.TrimPrefix(pkg, longest)
}

func formatTypeName(typName typesystem.TypeName) string {
	pkg := typName.Package
	name := typName.Name
	args := ""

	pkg = rewritePackageName(pkg)
	pkg = strings.ReplaceAll(pkg, "/", ".")

	if len(typName.Parameters) > 0 {
		args = "["
		for i, param := range typName.Parameters {
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
		typ: typ,

		vtables: map[string]*VTable{},
	}

	nt.def.Class = NodeClassGeneric
	nt.def.Name = formatTypeName(typ.Name())

	for _, opt := range options {
		opt(nt)
	}

	return nt
}
