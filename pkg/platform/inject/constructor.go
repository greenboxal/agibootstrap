package inject

import "reflect"

type constructorDefinition struct {
	parameters    []ServiceKey
	resultIndex   int
	errorOutIndex int
}

func Provide[T any](factory any) ServiceDefinition {
	definedType := reflect.TypeOf((*T)(nil)).Elem()
	constructorValue := reflect.ValueOf(factory)
	constructorType := constructorValue.Type()

	if constructorValue.Kind() != reflect.Func {
		panic("factory must be a function")
	}

	def := constructorDefinition{
		errorOutIndex: -1,
		resultIndex:   -1,
		parameters:    make([]ServiceKey, constructorType.NumIn()),
	}

	for i := 0; i < constructorType.NumIn(); i++ {
		def.parameters[i] = ServiceKeyFor(constructorType.In(i))
	}

	for i := 0; i < constructorType.NumOut(); i++ {
		arg := constructorType.Out(i)

		if arg.AssignableTo(errorType) {
			if i != constructorType.NumOut()-1 {
				panic("error must be the last return value")
			}

			if def.errorOutIndex != -1 {
				panic("multiple error return values")
			}

			def.errorOutIndex = i
		} else if arg.AssignableTo(definedType) {
			if def.resultIndex != -1 {
				panic("multiple return values assignable to T")
			}

			def.resultIndex = i
		} else {
			panic("return value not assignable to T or error")
		}
	}

	return ProvideFactory[T](func(ctx ResolutionContext) (empty T, _ error) {
		args := make([]reflect.Value, len(def.parameters))

		for i := 0; i < len(args); i++ {
			parameterType := def.parameters[i]

			if parameterType.Type == resolutionContextType {
				args[i] = reflect.ValueOf(ctx)
				continue
			}

			arg, err := ctx.GetService(parameterType)

			if err != nil {
				return empty, err
			}

			args[i] = reflect.ValueOf(arg)
		}

		result := constructorValue.Call(args)

		if def.errorOutIndex != -1 {
			errValue := result[def.errorOutIndex]

			if !errValue.IsNil() {
				return empty, errValue.Interface().(error)
			}
		}

		return result[0].Interface().(T), nil
	})
}
