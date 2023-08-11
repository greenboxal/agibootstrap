package psi

import (
	"context"
	"fmt"
	"reflect"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

type NodeInterface interface {
	Name() string
	Actions() []NodeActionDefinition

	ValidateImplementation(def VTableDefinition) error
}

type NodeActions interface {
	Invoke(ctx context.Context, name string, node Node, request any) (any, error)
}

type NodeInterfaceDefinition struct {
	Name    string                 `json:"name"`
	Actions []NodeActionDefinition `json:"actions"`
}

type NodeInterfaceOption func(*NodeInterfaceDefinition)

func WithInterfaceName(name string) NodeInterfaceOption {
	return func(ni *NodeInterfaceDefinition) {
		ni.Name = name
	}
}

func WithInterfaceAction(action NodeActionDefinition) NodeInterfaceOption {
	return func(ni *NodeInterfaceDefinition) {
		ni.Actions = append(ni.Actions, NodeActionDefinition{
			Name: "interface",
		})
	}
}

func WithInterfaceActions(actions ...NodeActionDefinition) NodeInterfaceOption {
	return func(ni *NodeInterfaceDefinition) {
		ni.Actions = append(ni.Actions, actions...)
	}
}

func DefineNodeInterface[T interface{}](options ...NodeInterfaceOption) NodeInterface {
	typ := reflect.TypeOf((*T)(nil)).Elem()

	return ReflectNodeInterface(typ, options...)
}

func ReflectNodeInterface(typ reflect.Type, options ...NodeInterfaceOption) NodeInterface {
	def := NodeInterfaceDefinition{
		Name: typ.Name(),
	}

	for _, opt := range options {
		opt(&def)
	}

	for i := 0; i < typ.NumMethod(); i++ {
		m := typ.Method(i)

		if !m.IsExported() {
			continue
		}

		ifaceAction := NodeActionDefinition{
			Name: m.Name,
		}

		if m.Type.NumIn() != 3 {
			panic(fmt.Errorf("method %s has %d parameters, expected 3", m.Name, m.Type.NumIn()))
		}

		if m.Type.NumOut() != 2 && m.Type.NumOut() != 1 {
			panic(fmt.Errorf("method %s has %d return values, expected 1 or 2", m.Name, m.Type.NumOut()))
		}

		requestType := reflect.PtrTo(m.Type.In(2))
		ifaceAction.RequestType = typesystem.TypeFrom(requestType)

		if m.Type.NumOut() > 1 {
			responseType := m.Type.Out(0)
			ifaceAction.ResponseType = typesystem.TypeFrom(responseType)
		}

		def.Actions = append(def.Actions, ifaceAction)
	}

	return &nodeInterface{
		definition: def,
	}
}

func NewNodeInterface(name string, options ...NodeInterfaceOption) NodeInterface {
	def := NodeInterfaceDefinition{
		Name: name,
	}

	for _, opt := range options {
		opt(&def)
	}

	return &nodeInterface{
		definition: def,
	}
}

type nodeInterface struct {
	definition NodeInterfaceDefinition
}

func (ni *nodeInterface) Name() string                    { return ni.definition.Name }
func (ni *nodeInterface) Actions() []NodeActionDefinition { return ni.definition.Actions }

func (ni *nodeInterface) ValidateImplementation(def VTableDefinition) error {
	for _, adef := range ni.definition.Actions {
		aimpl, ok := def.actions[adef.Name]

		if !ok {
			return fmt.Errorf("missing action %s", adef.Name)
		}

		if err := adef.ValidateImplementation(aimpl); err != nil {
			return fmt.Errorf("invalid action %s: %w", adef.Name, err)
		}
	}

	return nil
}

type VTable struct {
	iface NodeInterface
	def   VTableDefinition
}

func (t VTable) Name() string                    { return t.iface.Name() }
func (t VTable) Interface() NodeInterface        { return t.iface }
func (t VTable) Action(action string) NodeAction { return t.def.actions[action] }

type VTableDefinition struct {
	actions map[string]NodeAction
}

func BindInterface(iface NodeInterface, vtable VTableDefinition) *VTable {
	if err := iface.ValidateImplementation(vtable); err != nil {
		panic(err)
	}

	return &VTable{
		iface: iface,
		def:   vtable,
	}
}

func BindInterfaceFromNode(iface NodeInterface, typ typesystem.Type) *VTable {
	var def VTableDefinition

	def.actions = make(map[string]NodeAction)

	rt := typ.RuntimeType()

	for i, action := range iface.Actions() {
		m, ok := rt.MethodByName(action.Name)

		if !ok && i == 0 {
			rt = reflect.PtrTo(rt)
			m, ok = rt.MethodByName(action.Name)
		}

		if !ok {
			panic(fmt.Errorf("missing method %s", action.Name))
		}

		payloadTyp := m.Type.In(2)

		def.actions[action.Name] = &nodeAction{
			definition: action,

			handler: NodeActionFunc[Node, any, any](func(ctx context.Context, node Node, request any) (any, error) {
				vctx := reflect.ValueOf(ctx)
				vn := reflect.ValueOf(node)
				vreq := reflect.ValueOf(request)

				if payloadTyp.Kind() == reflect.Ptr {
					if vreq.CanAddr() {
						vreq = vreq.Addr()
					} else {
						vreq = reflect.New(payloadTyp.Elem())
						vreq.Elem().Set(reflect.ValueOf(request))
					}
				}

				vn.Method(m.Index).Call([]reflect.Value{
					vctx,
					vreq,
				})

				return nil, nil
			}),
		}
	}

	return BindInterface(iface, def)
}
