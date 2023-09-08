package psi

import (
	"context"
	"fmt"
	"reflect"

	"golang.org/x/exp/maps"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
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

		if m.Type.NumIn() < 1 && m.Type.NumIn() > 3 {
			continue
		}

		if m.Type.NumOut() != 2 && m.Type.NumOut() != 1 {
			continue
		}

		for i := 0; i < m.Type.NumIn(); i++ {
			t := m.Type.In(i)

			if t == contextType {
				continue
			}

			if t.AssignableTo(nodeRuntimeType) && ifaceAction.RequestType == nil && i < m.Type.NumIn()-1 {
				continue
			}

			if ifaceAction.RequestType == nil {
				ifaceAction.RequestType = typesystem.TypeFrom(t)
			}
		}

		if m.Type.NumOut() >= 1 {
			responseType := m.Type.Out(0)

			if !responseType.AssignableTo(errorType) {
				ifaceAction.ResponseType = typesystem.TypeFrom(responseType)
			}
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

func (vt VTableDefinition) Actions() []NodeAction { return maps.Values(vt.actions) }

func (vt VTableDefinition) Action(name string) NodeAction {
	return vt.actions[name]
}

func MakeVTableDefinition(actions map[string]NodeAction) VTableDefinition {
	return VTableDefinition{
		actions: actions,
	}
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

var errorType = reflect.TypeOf((*error)(nil)).Elem()
var nodeRuntimeType = reflect.TypeOf((*Node)(nil)).Elem()
var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

func BindInterfaceFromNode(iface NodeInterface, typ typesystem.Type) *VTable {
	var def VTableDefinition

	def.actions = make(map[string]NodeAction)

	rt := typ.RuntimeType()

	for i, action := range iface.Actions() {
		var payloadTyp *reflect.Type

		ctxIndex := -1
		selfIndex := -1
		payloadIndex := -1

		action := action

		m, ok := rt.MethodByName(action.Name)

		if !ok && i == 0 {
			rt = reflect.PtrTo(rt)
			m, ok = rt.MethodByName(action.Name)
		}

		if !ok {
			panic(fmt.Errorf("missing method %s", action.Name))
		}

		for i := 0; i < m.Type.NumIn(); i++ {
			t := m.Type.In(i)

			if t == contextType {
				ctxIndex = i
				continue
			}

			if t.AssignableTo(nodeRuntimeType) && payloadTyp == nil {
				selfIndex = i
				continue
			}

			if payloadTyp == nil {
				payloadIndex = i
				payloadTyp = &t
			}
		}

		def.actions[action.Name] = &nodeAction{
			definition: action,

			handler: NodeActionFunc[Node, any, any](func(ctx context.Context, node Node, request any) (any, error) {
				vctx := reflect.ValueOf(ctx)
				vn := reflect.ValueOf(node)
				vreq := reflect.ValueOf(request)
				argArr := make([]reflect.Value, m.Type.NumIn())

				if ctxIndex != -1 {
					argArr[ctxIndex] = vctx
				}

				if selfIndex != -1 {
					argArr[selfIndex] = vn
				}

				if payloadIndex != -1 && payloadTyp != nil && request != nil {
					if (*payloadTyp).Kind() == reflect.Ptr {
						if vreq.CanAddr() {
							vreq = vreq.Addr()
						} else {
							vreq = reflect.New((*payloadTyp).Elem())
							vreq.Elem().Set(reflect.ValueOf(request))
						}
					}

					argArr[payloadIndex] = vreq
				}

				r := m.Func.Call(argArr)

				if len(r) == 0 {
					return nil, nil
				}

				if len(r) == 1 {
					if r[0].IsNil() {
						return nil, nil
					}

					if r[0].Type().AssignableTo(errorType) {
						return nil, r[0].Interface().(error)
					}

					return r[0].Interface(), nil
				} else if len(r) == 2 {
					if r[1].IsNil() {
						return r[0].Interface(), nil
					}

					return nil, r[1].Interface().(error)
				}

				return nil, nil
			}),
		}
	}

	return BindInterface(iface, def)
}
