package inject

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

type ServiceProvider interface {
	ServiceLocator

	RegisterService(registration ServiceDefinition)
	Close(ctx context.Context) error

	AppendShutdownHook(func(ctx context.Context) error)
}

type ServiceRegistration interface {
	GetKey() ServiceKey
	GetDefinition() ServiceDefinition
	GetInstance(ctx ResolutionContext) (any, error)
}

type ResolutionContext interface {
	ServiceLocator

	Path() []ServiceKey

	AppendShutdownHook(func(ctx context.Context) error)
}

type CloseContext interface {
	Close(ctx context.Context) error
}

type ShutdownContext interface {
	Stop(ctx context.Context) error
}

var ErrServiceClosed = errors.New("service closed")
var ErrServiceNotFound = errors.New("service not found")
var ErrServiceDefinitionNotFound = errors.New("service definition not found")

func Inject[T any](sp ServiceLocator) T {
	result, err := sp.GetService(ServiceKeyOf[T]())

	if err != nil {
		panic(err)
	}

	return result.(T)
}

func RegisterInstance[T any](sp ServiceProvider, instance T) {
	sp.RegisterService(ProvideInstance[T](instance))
}

func ProvideInstance[T any](instance T) ServiceDefinition {
	return ProvideFactory[T](func(ctx ResolutionContext) (T, error) {
		return instance, nil
	})
}

func Register[T any](sp ServiceProvider, factory any) {
	sp.RegisterService(Provide[T](factory))
}

func RegisterFactory[T any](sp ServiceProvider, factory func(ctx ResolutionContext) (T, error)) {
	sp.RegisterService(ProvideFactory[T](factory))
}

func ProvideFactory[T any](factory func(ctx ResolutionContext) (T, error)) ServiceDefinition {
	return ServiceDefinition{
		Key:          ServiceKeyOf[T](),
		Dependencies: []ServiceKey{},
		Factory: func(ctx ResolutionContext, deps []any) (any, error) {
			return factory(ctx)
		},
	}
}

func ServiceKeyOf[T any]() ServiceKey {
	return ServiceKeyFor(reflect.TypeOf((*T)(nil)).Elem())
}

func ServiceKeyFor(typ reflect.Type) ServiceKey {
	realTyp := typ

	for realTyp.Kind() == reflect.Ptr {
		realTyp = realTyp.Elem()
	}

	return ServiceKey{
		Name: realTyp.PkgPath() + "/" + realTyp.Name(),
		Type: typ,
	}
}
