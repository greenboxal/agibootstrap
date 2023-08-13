package vm

import (
	"encoding/json"
	"strings"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"
	"rogchap.com/v8go"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

type LiveModule struct {
	ctx *Context
	mod *CachedModule

	instance *v8go.Object
	exports  *v8go.Object
	require  *v8go.Function

	loaded  bool
	loading bool
}

func NewLiveModule(ctx *Context, mod *CachedModule) (*LiveModule, error) {
	lm := &LiveModule{
		ctx: ctx,
		mod: mod,
	}

	if err := lm.initialize(); err != nil {
		return nil, err
	}

	return lm, nil
}

func (lm *LiveModule) register() error {
	_, err := lm.exports.MethodCall("register")

	return err
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

	result, err := lm.exports.MethodCall(name, castArgs...)

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
