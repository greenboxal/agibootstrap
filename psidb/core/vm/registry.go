package vm

import (
	"context"
	"sync"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/typesystem"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/services/typing"
)

type TypeRegistry struct {
	manager *typing.Manager
	parent  psi.TypeRegistry

	mu sync.RWMutex

	nodeTypeCache map[string]psi.NodeType
}

func NewTypeRegistry(manager *typing.Manager) *TypeRegistry {
	return &TypeRegistry{
		manager: manager,

		parent:        psi.GlobalTypeRegistry(),
		nodeTypeCache: make(map[string]psi.NodeType),
	}
}

func (tr *TypeRegistry) NodeTypes() []psi.NodeType { return tr.parent.NodeTypes() }
func (tr *TypeRegistry) EdgeTypes() []psi.EdgeType { return tr.EdgeTypes() }

func (tr *TypeRegistry) RegisterNodeType(nt psi.NodeType) {
	panic("not supported")
}

func (tr *TypeRegistry) RegisterEdgeKind(kind psi.EdgeKind, options ...psi.EdgeTypeOption) {
	panic("not supported")
}

func (tr *TypeRegistry) ReflectNodeType(typ typesystem.Type) psi.NodeType {
	return tr.parent.ReflectNodeType(typ)
}

func (tr *TypeRegistry) LookupEdgeType(kind psi.EdgeKind) psi.EdgeType {
	return tr.LookupEdgeType(kind)
}

func (tr *TypeRegistry) NodeTypeForType(ctx context.Context, typ *typing.Type) psi.NodeType {
	tr.mu.RLock()
	if nt := tr.nodeTypeCache[typ.FullName]; nt != nil {
		tr.mu.RUnlock()
		return nt
	}
	tr.mu.RUnlock()

	if nt := tr.parent.NodeTypeByName(ctx, typ.FullName); nt != nil {
		return nt
	}

	nt, err := NewDynamicType(ctx, tr, typ)

	if err != nil {
		panic(err)
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	if nt := tr.nodeTypeCache[typ.FullName]; nt != nil {
		return nt
	}

	tr.nodeTypeCache[typ.FullName] = nt

	return nt
}

func (tr *TypeRegistry) NodeTypeByPath(ctx context.Context, path psi.Path) (psi.NodeType, error) {
	tx := coreapi.GetTransaction(ctx)
	typ, err := psi.Resolve[*typing.Type](ctx, tx.Graph(), path)

	if err != nil {
		return nil, err
	}

	return tr.NodeTypeForType(ctx, typ), nil
}

func (tr *TypeRegistry) NodeTypeByName(ctx context.Context, name string) psi.NodeType {
	tr.mu.RLock()
	if nt := tr.nodeTypeCache[name]; nt != nil {
		tr.mu.RUnlock()
		return nt
	}
	tr.mu.RUnlock()

	nt := tr.parent.NodeTypeByName(ctx, name)

	if nt == nil {
		n, err := tr.lookupDynamicType(ctx, name)

		if err != nil {
			panic(err)
		}

		nt = n
	}

	if nt != nil {
		tr.mu.Lock()
		tr.nodeTypeCache[name] = nt
		tr.mu.Unlock()
	}

	return nt
}

func (tr *TypeRegistry) lookupDynamicType(ctx context.Context, name string) (psi.NodeType, error) {
	tx := coreapi.GetTransaction(ctx)

	if tx == nil {
		return nil, nil
	}

	definition, err := tr.manager.LookupType(ctx, name)

	if err != nil {
		return nil, err
	}

	if definition == nil {
		return nil, nil
	}

	return NewDynamicType(ctx, tr, definition)
}
