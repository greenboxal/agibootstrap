package vm

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/ipld/go-ipld-prime"
	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/typing"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type dynamicType struct {
	name   typesystem.TypeName
	typ    *typing.Type
	def    psi.NodeTypeDefinition
	ifaces map[string]*psi.VTable
}

func (d *dynamicType) OnAfterNodeLoaded(ctx context.Context, n psi.Node) error {
	return nil
}

func (d *dynamicType) OnBeforeNodeSaved(ctx context.Context, n psi.Node) error {
	return nil
}

func (d *dynamicType) EncodeNode(w io.Writer, encoder ipld.Encoder, n psi.Node) error {
	panic("implement me")
}

func (d *dynamicType) DecodeNode(r io.Reader, decoder ipld.Decoder) (psi.Node, error) {
	//TODO implement me
	panic("implement me")
}

func NewDynamicType(ctx context.Context, registry *TypeRegistry, definition *typing.Type) (psi.NodeType, error) {
	dt := &dynamicType{
		typ:  definition,
		name: typesystem.ParseTypeName(definition.FullName),

		def: psi.NodeTypeDefinition{
			Name: definition.FullName,
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
			act, err := newDynamicAction(ctx, registry, iface, action)

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
func (d *dynamicType) Type() typesystem.Type              { return VirtualNodeType }
func (d *dynamicType) TypeName() typesystem.TypeName      { return d.name }
func (d *dynamicType) RuntimeType() reflect.Type          { return VirtualNodeType.RuntimeType() }
func (d *dynamicType) Definition() psi.NodeTypeDefinition { return d.def }

func (d *dynamicType) Interfaces() []*psi.VTable         { return maps.Values(d.ifaces) }
func (d *dynamicType) Interface(name string) *psi.VTable { return d.ifaces[name] }

func (d *dynamicType) CreateInstance() psi.Node {
	return &VirtualNode{}
}

func (d *dynamicType) InitializeNode(n psi.Node) {
	n.(*VirtualNode).Init(n, psi.WithNodeType(d))
}

func (d *dynamicType) String() string {
	return d.def.Name
}

func newDynamicInterface(ctx context.Context, registry *TypeRegistry, def typing.InterfaceDefinition) (*dynamicInterface, error) {
	id := psi.NodeInterfaceDefinition{
		Name:        def.Name,
		Description: def.Description,
	}

	for _, action := range def.Actions {
		da := psi.NodeActionDefinition{
			Name:         action.Name,
			Description:  action.Description,
			RequestType:  nil,
			ResponseType: nil,
		}

		if action.RequestType != nil && !action.RequestType.IsEmpty() {
			requestType, err := registry.NodeTypeByPath(ctx, *action.RequestType)

			if err != nil {
				return nil, err
			}

			da.RequestType = requestType.Type()
		}

		if action.ResponseType != nil && !action.ResponseType.IsEmpty() {
			responseType, err := registry.NodeTypeByPath(ctx, *action.ResponseType)

			if err != nil {
				return nil, err
			}

			da.ResponseType = responseType.Type()
		}

		id.Actions = append(id.Actions, da)
	}

	return &dynamicInterface{
		definition:      def,
		ifaceDefinition: id,
	}, nil
}

type dynamicInterface struct {
	definition      typing.InterfaceDefinition
	ifaceDefinition psi.NodeInterfaceDefinition

	actions map[string]psi.NodeActionDefinition

	lm *ModuleInstance
}

func (d *dynamicInterface) Name() string                        { return d.ifaceDefinition.Name }
func (d *dynamicInterface) Description() string                 { return d.definition.Description }
func (d *dynamicInterface) Actions() []psi.NodeActionDefinition { return maps.Values(d.actions) }

func (d *dynamicInterface) ValidateImplementation(def psi.VTableDefinition) error {
	for _, adef := range d.ifaceDefinition.Actions {
		aimpl := def.Action(adef.Name)

		if err := adef.ValidateImplementation(aimpl); err != nil {
			return fmt.Errorf("invalid action %s: %w", adef.Name, err)
		}
	}

	return nil
}

func (d *dynamicInterface) getModule(ctx context.Context) (*ModuleInstance, error) {
	if d.lm != nil {
		return d.lm, nil
	}

	if d.definition.Module == nil {
		return nil, fmt.Errorf("interface %s has no module", d.Name())
	}

	tx := coreapi.GetTransaction(ctx)
	vmctx := inject.Inject[*Context](tx.Graph().Services())

	mod, err := psi.Resolve[*Module](ctx, tx.Graph(), *d.definition.Module)

	if err != nil {
		return nil, err
	}

	lm, err := vmctx.Load(ctx, mod)

	if err != nil {
		return nil, err
	}

	if _, err := lm.register(ctx, mod); err != nil {
		return nil, err
	}

	d.lm = lm

	return lm, nil
}

func newDynamicAction(ctx context.Context, registry *TypeRegistry, di *dynamicInterface, action typing.ActionDefinition) (psi.NodeAction, error) {
	da := &dynamicAction{
		di:  di,
		def: action,
	}

	if action.RequestType != nil && !action.RequestType.IsEmpty() {
		requestType, err := registry.NodeTypeByPath(ctx, *action.RequestType)

		if err != nil {
			return nil, err
		}

		da.requestType = requestType.Type()
	}

	if action.ResponseType != nil && !action.ResponseType.IsEmpty() {
		responseType, err := registry.NodeTypeByPath(ctx, *action.ResponseType)

		if err != nil {
			return nil, err
		}

		da.responseType = responseType.Type()
	}

	return da, nil
}

type dynamicAction struct {
	def          typing.ActionDefinition
	di           *dynamicInterface
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
	mod, err := d.di.getModule(ctx)

	if err != nil {
		return nil, err
	}

	var args []typesystem.Value

	if node != nil {
		args = append(args, typesystem.ValueOf(node))
	}

	if request != nil {
		args = append(args, typesystem.ValueOf(request))
	}

	r, err := mod.Invoke(d.def.BoundFunction, args...)

	if err != nil {
		return nil, err
	}

	return r.Value(), nil
}
