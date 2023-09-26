package vm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"rogchap.com/v8go"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type Context struct {
	baseCtx context.Context

	iso *Isolate
	sp  inject.ServiceLocator

	ctx *v8go.Context

	modules map[string]*ModuleInstance
	handles *ObjectHandleTable

	objKeysFn *v8go.Function

	console basicConsole
	timers  basicTimers
}

func NewContext(baseCtx context.Context, iso *Isolate, sp inject.ServiceLocator) *Context {
	ctx := &Context{
		baseCtx: baseCtx,
		iso:     iso,
		sp:      sp,
		ctx:     v8go.NewContext(iso.iso),

		modules: map[string]*ModuleInstance{},
		handles: NewObjectHandleTable(),
	}

	iso.registerContext(ctx)

	ctx.console = basicConsole{
		logger: logging.GetLoggerCtx(baseCtx, "vm/console").Desugar().ZapLogger().Sugar(),
	}

	ctx.timers = basicTimers{
		ctx:    ctx,
		timers: map[int]func(){},
	}

	MustSet(ctx.ctx.Global(), "global", ctx.ctx.Global())
	MustSet(ctx.ctx.Global(), "console", ctx.MustWrapValue(reflect.ValueOf(&ctx.console)))

	timers := ctx.MustWrapValue(reflect.ValueOf(&ctx.timers))

	MustSet(ctx.ctx.Global(), "__timers", timers)
	MustSet(ctx.ctx.Global(), "setTimeout", MustGet(timers.Object(), "setTimeout"))
	MustSet(ctx.ctx.Global(), "setInterval", MustGet(timers.Object(), "setInterval"))
	MustSet(ctx.ctx.Global(), "clearTimeout", MustGet(timers.Object(), "clearTimeout"))
	MustSet(ctx.ctx.Global(), "clearInterval", MustGet(timers.Object(), "clearInterval"))

	return ctx
}

func (vmctx *Context) Isolate() *Isolate      { return vmctx.iso }
func (vmctx *Context) Context() *v8go.Context { return vmctx.ctx }
func (vmctx *Context) Global() *v8go.Object   { return vmctx.ctx.Global() }

func (vmctx *Context) Load(ctx context.Context, m *Module) (*ModuleInstance, error) {
	if m.cached == nil {
		src := ModuleSource{
			Name:   m.Name,
			Source: m.Source,
		}

		if m.Source == "" && m.SourceFile != "" {
			data, err := os.ReadFile(m.SourceFile)

			if err != nil {
				return nil, err
			}

			src.Source = string(data)
		}

		src.Source = fmt.Sprintf("(function(module, exports, require) {%s\n})", src.Source)

		cached, err := vmctx.iso.moduleCache.Get(src)

		if err != nil {
			return nil, err
		}

		if cached == nil {
			cached = NewCachedModule(vmctx.iso.supervisor, src)
		}

		vmctx.iso.moduleCache.Add(cached)

		m.cached = cached
	}

	lm, err := NewLiveModule(vmctx, m.cached)

	if err != nil {
		return nil, err
	}

	vmctx.modules[m.Name] = lm

	return lm.Get()
}

func (vmctx *Context) Require(ctx context.Context, name string) (*ModuleInstance, error) {
	if m := vmctx.modules[name]; m != nil {
		return m.Get()
	}

	g := inject.Inject[psi.Graph](vmctx.sp)
	path, err := psi.ParsePath(name)

	if err != nil {
		return nil, err
	}

	mod, err := psi.Resolve[*Module](ctx, g, path)

	if err != nil {
		return nil, err
	}

	return mod.Get(ctx)
}

func (vmctx *Context) ConvertToArray(value *v8go.Value) ([]*v8go.Value, error) {
	var items []*v8go.Value

	obj := value.Object()

	for i := uint32(0); obj.HasIdx(i); i++ {
		value, err := obj.GetIdx(i)

		if err != nil {
			return nil, err
		}

		items = append(items, value)
	}

	return items, nil
}

func (vmctx *Context) GetObjectAsMap(value *v8go.Value) (map[string]*v8go.Value, error) {
	obj := value.Object()
	keysValue, err := vmctx.getObjectKeysFn().Call(obj, obj)

	if err != nil {
		return nil, err
	}

	keysObj := keysValue.Object()
	kv := map[string]*v8go.Value{}

	for i := uint32(0); keysObj.HasIdx(i); i++ {
		keyValue, err := keysObj.GetIdx(i)

		if err != nil {
			return nil, err
		}

		k := keyValue.String()
		v, err := obj.Get(k)

		if err != nil {
			return nil, err
		}

		kv[k] = v
	}

	return kv, nil
}

func (vmctx *Context) getObjectKeysFn() *v8go.Function {
	if vmctx.objKeysFn != nil {
		return vmctx.objKeysFn
	}

	objKeys, err := vmctx.ctx.RunScript("Object.keys", "")

	if err != nil {
		panic(err)
	}

	objKeysFn, err := objKeys.AsFunction()

	if err != nil {
		panic(err)
	}

	vmctx.objKeysFn = objKeysFn

	return objKeysFn
}

func (vmctx *Context) Close() error {
	vmctx.ctx.Close()

	vmctx.iso.unregisterContext(vmctx)

	return nil
}

func (vmctx *Context) MustWrapValue(value reflect.Value) *v8go.Value {
	wrapped, err := vmctx.WrapValue(value)

	if err != nil {
		panic(err)
	}

	return wrapped
}

func (vmctx *Context) WrapValue(value reflect.Value) (*v8go.Value, error) {
	if !value.IsValid() {
		return v8go.Undefined(vmctx.iso.iso), nil
	}

	if value.IsNil() {
		return v8go.Null(vmctx.iso.iso), nil
	}

	if value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface {
		proto := vmctx.iso.prototypeCache.Get(typesystem.TypeFrom(value.Type()))

		return proto.Wrap(vmctx, value)
	} else if value.Kind() == reflect.Struct {
		data, err := ipld.Encode(typesystem.ValueFrom(value).AsNode(), dagjson.Encode)

		if err != nil {
			return nil, err
		}

		return v8go.JSONParse(vmctx.ctx, string(data))
	}

	return v8go.NewValue(vmctx.iso.iso, value.Interface())
}

func (vmctx *Context) UnwrapValue(value *v8go.Value, dst reflect.Value) error {
	if value.IsUndefined() || value.IsNull() {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	switch dst.Kind() {
	case reflect.Struct:
		fallthrough
	case reflect.Ptr:
		if dst.Kind() == reflect.Struct || dst.Type().Elem().Kind() == reflect.Struct {
			data, err := v8go.JSONStringify(vmctx.ctx, value)

			if err != nil {
				return err
			}

			return json.Unmarshal([]byte(data), dst.Addr().Interface())
		}

		fallthrough

	case reflect.Interface:
		obj := value.Object()

		if obj.InternalFieldCount() == 2 {
			magic := obj.GetInternalField(0)

			if magic.IsUint32() && magic.Uint32() == handleMagicNumber {
				handle := obj.GetInternalField(1)
				v := vmctx.handles.Get(uint64(handle.Uint32()))

				dst.Set(v)
			}
		} else if dst.Type() == anyType {
			return vmctx.UnwrapValueAsAny(value, dst)
		} else {
			return errors.New("unsupported type")
		}

	case reflect.Slice:
		fallthrough
	case reflect.Array:
		items, err := vmctx.ConvertToArray(value)

		if err != nil {
			panic(err)
		}

		arr := reflect.MakeSlice(dst.Type(), len(items), len(items))

		for i, item := range items {
			err := vmctx.UnwrapValue(item, arr.Index(i))

			if err != nil {
				return err
			}
		}

	case reflect.Map:
		kv, err := vmctx.GetObjectAsMap(value)

		if err != nil {
			return err
		}

		m := reflect.MakeMap(dst.Type())

		for k, v := range kv {
			kv := reflect.New(dst.Type().Elem())

			if err := vmctx.UnwrapValue(v, kv.Elem()); err != nil {
				return err
			}

			m.SetMapIndex(reflect.ValueOf(k), kv.Elem())
		}

	case reflect.String:
		dst.SetString(value.String())
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		if value.IsInt32() {
			dst.SetInt(int64(value.Int32()))
		} else if value.IsUint32() {
			dst.SetInt(int64(int32(value.Uint32())))
		} else if value.IsNumber() {
			dst.SetInt(int64(value.Number()))
		} else if value.IsString() {
			i, err := strconv.ParseInt(value.String(), 10, 64)

			if err != nil {
				return err
			}

			dst.SetInt(i)
		} else {
			return errors.New("unsupported type")
		}
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		if value.IsInt32() {
			dst.SetUint(uint64(value.Int32()))
		} else if value.IsUint32() {
			dst.SetUint(uint64(value.Uint32()))
		} else if value.IsNumber() {
			dst.SetUint(uint64(value.Number()))
		} else if value.IsString() {
			i, err := strconv.ParseUint(value.String(), 10, 64)

			if err != nil {
				return err
			}

			dst.SetUint(i)
		} else {
			return errors.New("unsupported type")
		}
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		if value.IsInt32() {
			dst.SetFloat(float64(value.Int32()))
		} else if value.IsUint32() {
			dst.SetFloat(float64(value.Uint32()))
		} else if value.IsNumber() {
			dst.SetFloat(value.Number())
		} else if value.IsString() {
			f, err := strconv.ParseFloat(value.String(), 64)

			if err != nil {
				return err
			}

			dst.SetFloat(f)
		} else {
			return errors.New("unsupported type")
		}
	case reflect.Bool:
		if value.IsInt32() {
			dst.SetBool(value.Int32() != 0)
		} else if value.IsUint32() {
			dst.SetBool(value.Uint32() != 0)
		} else if value.IsNumber() {
			dst.SetBool(value.Number() != 0)
		} else if value.IsString() {
			dst.SetBool(len(value.String()) > 0)
		} else {
			dst.SetBool(value.IsNullOrUndefined())
		}

	default:
		return errors.New("unsupported type")
	}

	return nil
}

func (vmctx *Context) UnwrapValueAsAny(value *v8go.Value, elem reflect.Value) (err error) {
	var res any

	fallback := func() error {
		var v any

		data, err := v8go.JSONStringify(vmctx.ctx, value)

		if err != nil {
			return nil
		}

		if err := json.Unmarshal([]byte(data), &v); err != nil {
			return err
		}

		elem.Set(reflect.ValueOf(v))

		return nil
	}

	if value.IsNull() || value.IsUndefined() {
		elem.Set(reflect.Zero(elem.Type()))
		return nil
	}

	if value.IsNumber() {
		res = value.Number()
	} else if value.IsInt32() {
		res = value.Int32()
	} else if value.IsUint32() {
		res = value.Uint32()
	} else if value.IsString() {
		res = value.String()
	} else if value.IsBoolean() {
		res = value.Boolean()
	} else if value.IsArray() {
		res, err = vmctx.ConvertToArray(value)

		if err != nil {
			return err
		}
	} else if value.IsMap() {
		res, err = vmctx.GetObjectAsMap(value)

		if err != nil {
			return err
		}
	} else if value.IsObject() {
		obj := value.Object()

		if obj.InternalFieldCount() == 2 {
			magic := obj.GetInternalField(0)

			if magic.IsUint32() && magic.Uint32() == handleMagicNumber {
				handle := obj.GetInternalField(1)
				v := vmctx.handles.Get(uint64(handle.Uint32()))

				elem.Set(v)
				return nil
			}
		}

		return fallback()
	} else {
		return fallback()
	}

	elem.Set(reflect.ValueOf(res))

	return nil
}

func (vmctx *Context) Eval(source, origin string) (*v8go.Value, error) {
	result, err := vmctx.ctx.RunScript(source, origin)

	if err != nil {
		return nil, err
	}

	return result, nil
}
