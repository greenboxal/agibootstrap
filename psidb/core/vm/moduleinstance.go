package vm

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"rogchap.com/v8go"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/typing"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type ModuleInstance struct {
	ctx *Context
	mod *CachedModule

	instance *v8go.Object
	exports  *v8go.Object
	require  *v8go.Function

	loaded     bool
	loading    bool
	registered bool

	symbolMap map[string]*v8go.Value
}

func NewLiveModule(ctx *Context, mod *CachedModule) (*ModuleInstance, error) {
	lm := &ModuleInstance{
		ctx: ctx,
		mod: mod,

		symbolMap: map[string]*v8go.Value{},
	}

	if err := lm.initialize(); err != nil {
		return nil, err
	}

	return lm, nil
}

func (mi *ModuleInstance) register(ctx context.Context, self *Module) (any, error) {
	if mi.registered {
		return mi.symbolMap, nil
	}

	selfPath := self.CanonicalPath()

	tx := coreapi.GetTransaction(ctx)
	tm := inject.Inject[*typing.Manager](tx.Graph().Services())

	result, err := mi.exports.MethodCall("register")

	if err != nil {
		return nil, err
	}

	reg := func() error {

		moduleRegistration := result.Object()
		types, err := moduleRegistration.Get("types")

		if err != nil {
			return err
		}

		typesList := types.Object()

		for i := uint32(0); typesList.HasIdx(i); i++ {
			typeValue, err := typesList.GetIdx(i)

			if err != nil {
				return err
			}

			typeDef := typeValue.Object()
			typeNameValue, err := typeDef.Get("name")

			if err != nil {
				return err
			}

			typeFieldsValue, err := typeDef.Get("fields")

			if err != nil {
				return err
			}

			typeFields, err := mi.ctx.GetObjectAsMap(typeFieldsValue)

			if err != nil {
				return err
			}

			typeInterfacesValue, err := typeDef.Get("interfaces")

			if err != nil {
				return err
			}

			typeInterfaces, err := mi.ctx.GetObjectAsMap(typeInterfacesValue)

			if err != nil {
				return err
			}

			psiType := typing.NewType(typeNameValue.String())

			for _, fieldDefValue := range typeFields {
				fld := typing.FieldDefinition{}
				fieldDef := fieldDefValue.Object()

				fieldNameValue, err := fieldDef.Get("name")

				if err != nil {
					return err
				}

				fieldTypeValue, err := fieldDef.Get("type")

				if err != nil {
					return err
				}

				fld.Name = fieldNameValue.String()

				if fieldTypeValue.IsString() && fieldTypeValue.String() != "" {
					fld.Type, err = psi.ParsePath(fieldTypeValue.String())

					if err != nil {
						return err
					}
				}

				psiType.Fields = append(psiType.Fields, fld)
			}

			for vtableKey, vtableValue := range typeInterfaces {
				iface := typing.InterfaceDefinition{}
				iface.Module = &selfPath

				vtableDef := vtableValue.Object()

				actionsImplValue, err := vtableDef.Get("actions")

				if err != nil {
					return err
				}

				actionsImplMap, err := mi.ctx.GetObjectAsMap(actionsImplValue)

				if err != nil {
					return err
				}

				ifaceValue, err := vtableDef.Get("interface")

				if err != nil {
					return err
				}

				ifaceDef := ifaceValue.Object()
				ifaceNameValue, err := ifaceDef.Get("name")

				if err != nil {
					return err
				}

				iface.Name = ifaceNameValue.String()

				actionsValue, err := ifaceDef.Get("actions")

				if err != nil {
					return err
				}

				actions, err := mi.ctx.GetObjectAsMap(actionsValue)

				if err != nil {
					return err
				}

				for actionKey, actionValue := range actions {
					act := typing.ActionDefinition{}
					actionDef := actionValue.Object()

					actionNameValue, err := actionDef.Get("name")

					if err != nil {
						return err
					}

					actionRequestTypeValue, err := actionDef.Get("request_type")

					if err != nil {
						return err
					}

					actionResponseTypeValue, err := actionDef.Get("response_type")

					if err != nil {
						return err
					}

					act.Name = actionNameValue.String()

					if actionRequestTypeValue.IsString() && actionRequestTypeValue.String() != "" {
						t, err := psi.ParsePath(actionRequestTypeValue.String())

						if err != nil {
							return err
						}

						act.RequestType = &t
					}

					if actionResponseTypeValue.IsString() && actionResponseTypeValue.String() != "" {
						t, err := psi.ParsePath(actionResponseTypeValue.String())

						if err != nil {
							return err
						}

						act.ResponseType = &t
					}

					actionImpl := actionsImplMap[actionKey]

					if actionImpl != nil {
						act.BoundFunction = fmt.Sprintf("%s.%s", vtableKey, actionKey)

						mi.symbolMap[act.BoundFunction] = actionImpl
					}

					iface.Actions = append(iface.Actions, act)
				}

				psiType.Interfaces = append(psiType.Interfaces, iface)
			}

			if _, err := tm.CreateType(ctx, psiType); err != nil {
				return err
			}
		}

		return nil
	}

	if err := reg(); err != nil {
		return nil, err
	}

	mi.registered = true

	return result, nil
}

func (mi *ModuleInstance) dispatch() error {
	_, err := mi.exports.MethodCall("dispatch")

	return err
}

func (mi *ModuleInstance) initialize() error {
	var err error

	modTemplate := v8go.NewObjectTemplate(mi.ctx.iso.iso)
	exportsTemplate := v8go.NewObjectTemplate(mi.ctx.iso.iso)
	requireTemplate := v8go.NewFunctionTemplate(mi.ctx.iso.iso, mi.requireCallback)

	mi.instance, err = modTemplate.NewInstance(mi.ctx.ctx)

	if err != nil {
		return err
	}

	mi.exports, err = exportsTemplate.NewInstance(mi.ctx.ctx)

	if err != nil {
		return err
	}

	mi.require = requireTemplate.GetFunction(mi.ctx.ctx)

	if err := mi.instance.Set("exports", mi.exports); err != nil {
		return err
	}

	if err := mi.instance.Set("require", mi.require); err != nil {
		return err
	}

	return nil
}

func (mi *ModuleInstance) load() error {
	if mi.loading {
		return errors.New("module is already loading")
	}

	mi.loading = true

	defer func() {
		mi.loading = false
		mi.loaded = true
	}()

	script, err := mi.mod.GetCached(mi.ctx.iso)

	if err != nil {
		return err
	}

	v, err := script.Run(mi.ctx.ctx)

	if err != nil {
		return err
	}

	f, err := v.AsFunction()

	if err != nil {
		return err
	}

	if _, err := f.Call(f, mi.instance, mi.exports, mi.require); err != nil {
		return err
	}

	exports, err := mi.instance.Get("exports")

	if err != nil {
		return err
	}

	mi.exports, err = exports.AsObject()

	if err != nil {
		return err
	}

	return nil
}

func (mi *ModuleInstance) requireCallback(info *v8go.FunctionCallbackInfo) *v8go.Value {
	name := info.Args()[0].String()
	name = strings.TrimPrefix(name, "./")

	mod, err := mi.ctx.Require(mi.ctx.baseCtx, name)

	if err != nil {
		panic(err)
	}

	loaded, err := mod.Get()

	if err != nil {
		panic(err)
	}

	return loaded.exports.Value
}

func (mi *ModuleInstance) Get() (*ModuleInstance, error) {
	if !mi.loading && !mi.loaded {
		if err := mi.load(); err != nil {
			return nil, err
		}
	}

	return mi, nil
}

func (mi *ModuleInstance) Invoke(name string, args ...typesystem.Value) (typesystem.Value, error) {
	var err error
	var result *v8go.Value

	castArgs := make([]v8go.Valuer, len(args))

	for i, arg := range args {
		wrapped, err := mi.ctx.WrapValue(arg.Value())

		if err != nil {
			return typesystem.Value{}, err
		}

		castArgs[i] = wrapped
	}

	if sym := mi.symbolMap[name]; sym != nil {
		symFn, err := sym.AsFunction()

		if err != nil {
			return typesystem.Value{}, err
		}

		result, err = symFn.Call(symFn, castArgs...)
	} else {
		result, err = mi.exports.MethodCall(name, castArgs...)
	}

	if err != nil {
		return typesystem.Value{}, err
	}

	kv := reflect.New(reflect.TypeOf((*any)(nil)).Elem())

	if err := mi.ctx.UnwrapValue(result, kv.Elem()); err != nil {
		return typesystem.Value{}, err
	}

	return typesystem.ValueFrom(kv.Elem()), nil
}
