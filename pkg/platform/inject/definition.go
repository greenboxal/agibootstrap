package inject

import "reflect"

type ServiceKey struct {
	Name string
	Type reflect.Type
}

type ServiceFactory func(ctx ResolutionContext, deps []any) (any, error)

type ServiceDefinition struct {
	Key          ServiceKey
	Factory      ServiceFactory
	Dependencies []ServiceKey
}
