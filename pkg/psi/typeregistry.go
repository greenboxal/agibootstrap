package psi

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

var nodeTypeRegistryMutex sync.RWMutex
var nodeTypeRegistry = map[typesystem.Type]NodeType{}
var nodeTypeByName = map[string]NodeType{}

var edgeTypeRegistryMutex sync.RWMutex
var edgeTypeByName = map[EdgeKind]EdgeType{}

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
	nodeTypeRegistryMutex.RLock()
	defer nodeTypeRegistryMutex.RUnlock()

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

var packageTypeNameMap = map[string]string{
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs": "vfs",
	"github.com/greenboxal/agibootstrap/pkg/":             "agib.",
	"github.com/greenboxal/agibootstrap/psidb/db/":        "psidb.",
	"github.com/greenboxal/agibootstrap/psidb/services/":  "psidb.",
	"github.com/greenboxal/agibootstrap/psidb/modules/":   "",
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

func LookupEdgeType(kind EdgeKind) EdgeType {
	edgeTypeRegistryMutex.RLock()
	defer edgeTypeRegistryMutex.RUnlock()

	return edgeTypeByName[kind]
}

func DefineEdgeType[T Node](kind EdgeKind, options ...EdgeTypeOption) TypedEdgeKind[T] {
	edgeTypeRegistryMutex.Lock()
	defer edgeTypeRegistryMutex.Unlock()

	if _, ok := edgeTypeByName[kind]; ok {
		panic("edge type already defined")
	}

	et := &edgeType{}
	et.kind = kind
	et.name = string(kind)

	for _, opt := range options {
		opt(et)
	}

	edgeTypeByName[kind] = et

	return TypedEdgeKind[T](kind)
}
