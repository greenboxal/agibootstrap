package typing

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/greenboxal/aip/aip-sdk/pkg/utils"
	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/typesystem"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/vm"
)

type TypeRegistry struct {
	manager *Manager
	parent  psi.TypeRegistry

	mu            sync.RWMutex
	nodeTypeCache map[string]psi.NodeType
}

func NewTypeRegistry(manager *Manager) *TypeRegistry {
	return &TypeRegistry{
		manager: manager,

		parent:        psi.GlobalTypeRegistry(),
		nodeTypeCache: make(map[string]psi.NodeType),
	}
}

func (t *TypeRegistry) NodeTypes() []psi.NodeType { return t.parent.NodeTypes() }
func (t *TypeRegistry) EdgeTypes() []psi.EdgeType { return t.EdgeTypes() }

func (t *TypeRegistry) RegisterNodeType(nt psi.NodeType) {
	panic("not supported")
}

func (t *TypeRegistry) RegisterEdgeKind(kind psi.EdgeKind, options ...psi.EdgeTypeOption) {
	panic("not supported")
}

func (t *TypeRegistry) ReflectNodeType(typ typesystem.Type) psi.NodeType {
	return t.parent.ReflectNodeType(typ)
}

func (t *TypeRegistry) LookupEdgeType(kind psi.EdgeKind) psi.EdgeType {
	return t.LookupEdgeType(kind)
}

func (t *TypeRegistry) NodeTypeByName(ctx context.Context, name string) psi.NodeType {
	t.mu.Lock()
	defer t.mu.Unlock()

	if nt := t.nodeTypeCache[name]; nt != nil {
		return nt
	}

	nt := t.parent.NodeTypeByName(ctx, name)

	if nt == nil {
		n, err := t.lookupDynamicType(ctx, name)

		if err != nil {
			panic(err)
		}

		nt = n
	}

	if nt != nil {
		t.nodeTypeCache[name] = nt
	}

	return nt
}

func (t *TypeRegistry) lookupDynamicType(ctx context.Context, name string) (psi.NodeType, error) {
	tx := coreapi.GetTransaction(ctx)

	if tx == nil {
		return nil, nil
	}

	definition, err := t.manager.lookupType(ctx, name)

	if err != nil {
		return nil, err
	}

	if definition == nil {
		return nil, nil
	}

	return NewDynamicType(ctx, t, definition)
}

type opaqueNode struct {
	psi.NodeBase
}

var opaqueNodeType = typesystem.TypeOf((*opaqueNode)(nil))

type dynamicType struct {
	name   typesystem.TypeName
	typ    *Type
	def    psi.NodeTypeDefinition
	ifaces map[string]*psi.VTable
}

func NewDynamicType(ctx context.Context, registry *TypeRegistry, definition *Type) (psi.NodeType, error) {
	dt := &dynamicType{
		typ:  definition,
		name: typesystem.AsTypeName(utils.ParseTypeName(definition.Name)),

		def: psi.NodeTypeDefinition{
			Name: definition.Name,
		},

		ifaces: map[string]*psi.VTable{},
	}

	for _, def := range definition.Interfaces {
		actions := map[string]psi.NodeAction{}
		iface, err := newDynamicInterface(ctx, registry, def)

		if err != nil {
			return nil, err
		}

		for _, action := range def.Actions {
			act, err := newDynamicAction(ctx, registry, action)

			if err != nil {
				return nil, err
			}

			actions[action.Name] = act
		}

		vt := psi.MakeVTableDefinition(actions)
		dt.ifaces[def.Name] = psi.BindInterface(iface, vt)
	}

	return dt, nil
}

func (d *dynamicType) Name() string                       { return d.def.Name }
func (d *dynamicType) Type() typesystem.Type              { return opaqueNodeType }
func (d *dynamicType) TypeName() typesystem.TypeName      { return d.name }
func (d *dynamicType) RuntimeType() reflect.Type          { return opaqueNodeType.RuntimeType() }
func (d *dynamicType) Definition() psi.NodeTypeDefinition { return d.def }

func (d *dynamicType) Interfaces() []*psi.VTable         { return maps.Values(d.ifaces) }
func (d *dynamicType) Interface(name string) *psi.VTable { return d.ifaces[name] }

func (d *dynamicType) CreateInstance() psi.Node {
	return &opaqueNode{}
}

func (d *dynamicType) InitializeNode(n psi.Node) {
	n.(*opaqueNode).Init(n, psi.WithNodeType(d))
}

func (d *dynamicType) String() string {
	return d.def.Name
}

func newDynamicInterface(ctx context.Context, t *TypeRegistry, def InterfaceDefinition) (psi.NodeInterface, error) {
	id := psi.NodeInterfaceDefinition{
		Name: def.Name,
	}

	for _, action := range def.Actions {
		id.Actions = append(id.Actions, psi.NodeActionDefinition{
			Name:         action.Name,
			RequestType:  nil,
			ResponseType: nil,
		})
	}

	return &dynamicInterface{
		def: id,
	}, nil
}

type dynamicInterface struct {
	def     psi.NodeInterfaceDefinition
	actions map[string]psi.NodeActionDefinition
}

func (d *dynamicInterface) Name() string { return d.def.Name }

func (d *dynamicInterface) Actions() []psi.NodeActionDefinition { return maps.Values(d.actions) }

func (d *dynamicInterface) ValidateImplementation(def psi.VTableDefinition) error {
	for _, adef := range d.def.Actions {
		aimpl := def.Action(adef.Name)

		if err := adef.ValidateImplementation(aimpl); err != nil {
			return fmt.Errorf("invalid action %s: %w", adef.Name, err)
		}
	}

	return nil
}

func newDynamicAction(ctx context.Context, registry *TypeRegistry, action ActionDefinition) (psi.NodeAction, error) {
	return &dynamicAction{
		def: action,
	}, nil
}

type dynamicAction struct {
	def          ActionDefinition
	mod          *vm.LiveModule
	requestType  typesystem.Type
	responseType typesystem.Type
}

func (d *dynamicAction) Name() string {
	return d.def.Name
}

func (d *dynamicAction) RequestType() typesystem.Type {
	return d.requestType
}

func (d *dynamicAction) ResponseType() typesystem.Type {
	return d.responseType
}

func (d *dynamicAction) Invoke(ctx context.Context, node psi.Node, request any) (any, error) {
	r, err := d.mod.Invoke(d.def.BoundFunction, typesystem.ValueOf(node), typesystem.ValueOf(request))

	if err != nil {
		return nil, err
	}

	return r.Value(), nil
}
