package vm

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/stoewer/go-strcase"

	"rogchap.com/v8go"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type RuntimePrototypeCache struct {
	mu    sync.RWMutex
	iso   *Isolate
	cache map[typesystem.Type]*ObjectPrototype
}

func NewRuntimePrototypeCache(iso *Isolate) *RuntimePrototypeCache {
	return &RuntimePrototypeCache{
		iso:   iso,
		cache: map[typesystem.Type]*ObjectPrototype{},
	}
}

func (c *RuntimePrototypeCache) Get(t typesystem.Type) *ObjectPrototype {
	c.mu.RLock()
	if op := c.cache[t]; op != nil {
		c.mu.RUnlock()

		return op
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if op := c.cache[t]; op != nil {
		return op
	}

	op, err := PrototypeForType(c.iso, t)

	if err != nil {
		panic(err)
	}

	c.cache[t] = op

	return op
}

type ObjectPrototype struct {
	Template *v8go.ObjectTemplate
	Methods  []*PrototypeFunction
}

func (op *ObjectPrototype) Wrap(ctx *Context, v reflect.Value) (*v8go.Value, error) {
	if v.Kind() != reflect.Ptr && v.Kind() != reflect.Interface {
		if !v.CanAddr() {
			return nil, errors.New("value is not addressable")
		}

		v = v.Addr()
	}

	_, instance, err := ctx.handles.Add(v, func(handle int) (*v8go.Value, error) {
		instance, err := op.Template.NewInstance(ctx.ctx)

		if err != nil {
			return nil, err
		}

		if err := instance.SetInternalField(0, handleMagicNumber); err != nil {
			return nil, err
		}

		if err := instance.SetInternalField(1, uint32(handle)); err != nil {
			return nil, err
		}

		return instance.Value, nil
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}

func PrototypeForType(iso *Isolate, t typesystem.Type) (*ObjectPrototype, error) {
	op := &ObjectPrototype{}

	walkMethods := func(typ reflect.Type) error {
		for i := 0; i < typ.NumMethod(); i++ {
			m := typ.Method(i)

			if !m.IsExported() {
				continue
			}

			pf := ReflectFunctionPrototype(m)

			op.Methods = append(op.Methods, pf)
		}

		return nil
	}

	if err := walkMethods(t.RuntimeType()); err != nil {
		return nil, err
	}

	if err := walkMethods(reflect.PointerTo(t.RuntimeType())); err != nil {
		return nil, err
	}

	op.Template = v8go.NewObjectTemplate(iso.iso)
	op.Template.SetInternalFieldCount(2)

	for _, pf := range op.Methods {
		fn := pf.BuildFunctionTemplate(iso)
		mangledName := strcase.LowerCamelCase(pf.Method.Name)

		if err := op.Template.Set(mangledName, fn, v8go.DontDelete); err != nil {
			return nil, err
		}
	}

	return op, nil
}

type PrototypeFunction struct {
	Method reflect.Method

	IsVariadic         bool
	FirstArgumentIndex int
	FixedArgumentCount int
	ReturnValueCount   int

	InCtxIndex  int
	OutErrIndex int
}

func ReflectFunctionPrototype(m reflect.Method) *PrototypeFunction {
	var pf PrototypeFunction

	pf.Method = m
	pf.IsVariadic = m.Type.IsVariadic()
	pf.FirstArgumentIndex = 0
	pf.FixedArgumentCount = m.Type.NumIn() - 1
	pf.ReturnValueCount = m.Type.NumOut()

	pf.InCtxIndex = -1
	pf.OutErrIndex = -1

	if pf.IsVariadic {
		pf.FixedArgumentCount--
	}

	for i := 1; i < m.Type.NumIn(); i++ {
		if pf.InCtxIndex == -1 && m.Type.In(i) == contextType {
			pf.FixedArgumentCount--
			pf.FirstArgumentIndex++
			pf.InCtxIndex = i - 1
			break
		}
	}

	for i := 0; i < m.Type.NumOut(); i++ {
		if pf.OutErrIndex == -1 && m.Type.Out(i) == errorType {
			pf.OutErrIndex = i
			break
		}
	}

	if pf.OutErrIndex != -1 && pf.OutErrIndex != m.Type.NumOut()-1 {
		pf.OutErrIndex = -1
	} else if pf.OutErrIndex != -1 {
		pf.ReturnValueCount--
	}

	return &pf
}

func (pf PrototypeFunction) Call(ctx *Context, info *v8go.FunctionCallbackInfo) (v8Result *v8go.Value) {
	var panicValue any
	var goResultValues []reflect.Value

	defer func() {
		if err := recover(); err != nil {
			panicValue = err
		}

		if panicValue != nil {
			ex := fmt.Sprintf("%+#v", panicValue)
			exValue, err := v8go.NewValue(info.Context().Isolate(), ex)

			if err != nil {
				panic(err)
			}

			v8Result = info.Context().Isolate().ThrowException(exValue)
		}
	}()

	receiverHandle := info.This().GetInternalField(1)
	receiver := ctx.handles.Get(uint64(receiverHandle.Uint32()))
	fn := receiver.MethodByName(pf.Method.Name)

	vmArgs := info.Args()
	goArgs := make([]reflect.Value, fn.Type().NumIn())

	if pf.InCtxIndex != -1 {
		goArgs[pf.InCtxIndex] = reflect.ValueOf(ctx.baseCtx)
	}

	for i := 0; i < pf.FixedArgumentCount; i++ {
		v := reflect.New(fn.Type().In(pf.FirstArgumentIndex + i))

		if err := ctx.UnwrapValue(vmArgs[i], v.Elem()); err != nil {
			panic(err)
		}

		goArgs[pf.FirstArgumentIndex+i] = v.Elem()
	}

	if pf.IsVariadic {
		varArgs := reflect.MakeSlice(
			fn.Type().In(pf.FixedArgumentCount),
			len(vmArgs)-pf.FixedArgumentCount,
			len(vmArgs)-pf.FixedArgumentCount,
		)

		for i := pf.FixedArgumentCount; i < len(vmArgs); i++ {
			arg := varArgs.Index(i - pf.FixedArgumentCount)

			if err := ctx.UnwrapValue(vmArgs[i], arg); err != nil {
				panic(err)
			}
		}

		goArgs[len(goArgs)-1] = varArgs
	}

	func() {
		defer func() {
			if err := recover(); err != nil {
				panicValue = err
			}
		}()

		if pf.IsVariadic {
			goResultValues = fn.CallSlice(goArgs)
		} else {
			goResultValues = fn.Call(goArgs)
		}
	}()

	if panicValue != nil {
		return nil
	}

	if len(goResultValues) == 0 {
		return v8go.Undefined(info.Context().Isolate())
	}

	if pf.OutErrIndex != -1 {
		errValue := goResultValues[pf.OutErrIndex]

		if !errValue.IsNil() {
			panic(errValue.Interface().(error))
		}
	}

	if pf.ReturnValueCount == 0 {
		return v8go.Undefined(info.Context().Isolate())
	} else if pf.ReturnValueCount == 1 {
		v, err := ctx.WrapValue(goResultValues[0])

		if err != nil {
			panic(err)
		}

		return v
	} else {
		arr, err := v8go.NewArray(info.Context(), pf.ReturnValueCount)

		if err != nil {
			panic(err)
		}

		for i := 0; i < pf.ReturnValueCount; i++ {
			v, err := ctx.WrapValue(goResultValues[i])

			if err != nil {
				panic(err)
			}

			if err := arr.SetIdx(uint32(i), v); err != nil {
				panic(err)
			}
		}

		return arr.Value
	}
}

func (pf PrototypeFunction) BuildFunctionTemplate(iso *Isolate) *v8go.FunctionTemplate {
	return v8go.NewFunctionTemplate(iso.iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		ctx := iso.getContext(info.Context())

		return pf.Call(ctx, info)
	})
}
