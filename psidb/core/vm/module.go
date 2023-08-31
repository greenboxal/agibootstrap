package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"
	"rogchap.com/v8go"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/typesystem"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/services/typing"
)

type LiveModule struct {
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

func NewLiveModule(ctx *Context, mod *CachedModule) (*LiveModule, error) {
	lm := &LiveModule{
		ctx: ctx,
		mod: mod,

		symbolMap: map[string]*v8go.Value{},
	}

	if err := lm.initialize(); err != nil {
		return nil, err
	}

	return lm, nil
}

func (lm *LiveModule) register(ctx context.Context, self *Module) (any, error) {
	if lm.registered {
		return lm.symbolMap, nil
	}

	selfPath := self.CanonicalPath()

	tx := coreapi.GetTransaction(ctx)
	tm := inject.Inject[*typing.Manager](tx.Graph().Services())

	result, err := lm.exports.MethodCall("register")

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

			typeFields, err := lm.ctx.GetObjectAsMap(typeFieldsValue)

			if err != nil {
				return err
			}

			typeInterfacesValue, err := typeDef.Get("interfaces")

			if err != nil {
				return err
			}

			typeInterfaces, err := lm.ctx.GetObjectAsMap(typeInterfacesValue)

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

				actionsImplMap, err := lm.ctx.GetObjectAsMap(actionsImplValue)

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

				actions, err := lm.ctx.GetObjectAsMap(actionsValue)

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

						lm.symbolMap[act.BoundFunction] = actionImpl
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

	lm.registered = true

	return result, nil
}

func (lm *LiveModule) dispatch() error {
	_, err := lm.exports.MethodCall("dispatch")

	return err
}

func (lm *LiveModule) initialize() error {
	var err error

	modTemplate := v8go.NewObjectTemplate(lm.ctx.iso.iso)
	exportsTemplate := v8go.NewObjectTemplate(lm.ctx.iso.iso)
	requireTemplate := v8go.NewFunctionTemplate(lm.ctx.iso.iso, lm.requireCallback)

	lm.instance, err = modTemplate.NewInstance(lm.ctx.ctx)

	if err != nil {
		return err
	}

	lm.exports, err = exportsTemplate.NewInstance(lm.ctx.ctx)

	if err != nil {
		return err
	}

	lm.require = requireTemplate.GetFunction(lm.ctx.ctx)

	if err := lm.instance.Set("exports", lm.exports); err != nil {
		return err
	}

	if err := lm.instance.Set("require", lm.require); err != nil {
		return err
	}

	return nil
}

func (lm *LiveModule) load() error {
	if lm.loading {
		return errors.New("module is already loading")
	}

	lm.loading = true

	defer func() {
		lm.loading = false
		lm.loaded = true
	}()

	script, err := lm.mod.GetCached(lm.ctx.iso)

	if err != nil {
		return err
	}

	v, err := script.Run(lm.ctx.ctx)

	if err != nil {
		return err
	}

	f, err := v.AsFunction()

	if err != nil {
		return err
	}

	if _, err := f.Call(f, lm.instance, lm.exports, lm.require); err != nil {
		return err
	}

	exports, err := lm.instance.Get("exports")

	if err != nil {
		return err
	}

	lm.exports, err = exports.AsObject()

	if err != nil {
		return err
	}

	return nil
}

func (lm *LiveModule) requireCallback(info *v8go.FunctionCallbackInfo) *v8go.Value {
	name := info.Args()[0].String()
	name = strings.TrimPrefix(name, "./")

	mod, err := lm.ctx.Require(lm.ctx.baseCtx, name)

	if err != nil {
		panic(err)
	}

	loaded, err := mod.Get()

	if err != nil {
		panic(err)
	}

	return loaded.exports.Value
}

func (lm *LiveModule) Get() (*LiveModule, error) {
	if !lm.loading && !lm.loaded {
		if err := lm.load(); err != nil {
			return nil, err
		}
	}

	return lm, nil
}

func (lm *LiveModule) Invoke(name string, args ...typesystem.Value) (typesystem.Value, error) {
	var err error
	var result *v8go.Value

	castArgs := make([]v8go.Valuer, len(args))

	for i, arg := range args {
		encoded, err := ipld.Encode(typesystem.Wrap(arg.Value().Interface()), dagjson.Encode)

		if err != nil {
			return typesystem.Value{}, err
		}

		v, err := v8go.JSONParse(lm.ctx.ctx, string(encoded))

		if err != nil {
			return typesystem.Value{}, err
		}

		castArgs[i] = v
	}

	if sym := lm.symbolMap[name]; sym != nil {
		symFn, err := sym.AsFunction()

		if err != nil {
			return typesystem.Value{}, err
		}

		result, err = symFn.Call(symFn, castArgs...)
	} else {
		result, err = lm.exports.MethodCall(name, castArgs...)
	}

	if err != nil {
		return typesystem.Value{}, err
	}

	if result.IsNullOrUndefined() {
		return typesystem.Value{}, nil
	}

	encoded, err := v8go.JSONStringify(lm.ctx.ctx, result)

	if err != nil {
		return typesystem.Value{}, err
	}

	var kv map[string]interface{}

	if err := json.Unmarshal([]byte(encoded), &kv); err != nil {
		return typesystem.Value{}, err
	}

	return typesystem.ValueOf(kv), nil
}
