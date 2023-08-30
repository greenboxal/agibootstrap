package vm

import (
	"reflect"

	"rogchap.com/v8go"
)

func BindObject(iso *Isolate, obj any) *v8go.ObjectTemplate {
	v := reflect.ValueOf(obj)
	t := v.Type()

	structType := t

	for structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	tmpl := v8go.NewObjectTemplate(iso.iso)

	for i := 0; i < t.NumField(); i++ {
		field := structType.Field(i)

		if field.PkgPath != "" {
			continue
		}

		if err := tmpl.Set(field.Name, v.Field(i).Interface(), v8go.DontDelete); err != nil {
			panic(err)
		}
	}

	for i := 0; i < t.NumMethod(); i++ {
		fn := BindFunction(iso, v.Method(i).Interface())

		if err := tmpl.Set(t.Method(i).Name, fn); err != nil {
			panic(err)
		}
	}

	return tmpl
}

func BindFunction(iso *Isolate, fn any) *v8go.FunctionTemplate {
	v := reflect.ValueOf(fn)
	t := v.Type()

	return v8go.NewFunctionTemplate(iso.iso, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		args := make([]reflect.Value, t.NumIn())

		for i := 0; i < t.NumIn(); i++ {
			args[i] = ConvertJsToGo(iso, info.Args()[i])
		}

		result := v.Call(args)

		if len(result) == 0 {
			return v8go.Null(iso.iso)
		} else if len(result) == 2 {
			if err, ok := result[1].Interface().(error); ok {
				panic(err)
			}
		}

		return ConvertGoToJs(iso, result[0])
	})
}

func ConvertGoToJs(iso *Isolate, value reflect.Value) *v8go.Value {
	v, err := v8go.NewValue(iso.iso, value.Interface())

	if err != nil {
		panic(err)
	}

	return v
}

func ConvertJsToGo(iso *Isolate, value *v8go.Value) reflect.Value {
	panic("LOL")
}
