package inject

import "reflect"

type ServiceKey interface {
	RuntimeType() reflect.Type
	String() string
}

type serviceKey struct {
	name string
	typ  reflect.Type
}

func (k *serviceKey) RuntimeType() reflect.Type { return k.typ }
func (k *serviceKey) String() string            { return k.name }

type ServiceLocator interface {
	GetService(key ServiceKey) (any, error)
}

func buildKey[T any]() ServiceKey {
	typ := reflect.TypeOf((*T)(nil)).Elem()

	return &serviceKey{
		name: typ.Name(),
		typ:  typ,
	}
}

func GetService[T any](locator ServiceLocator) T {
	key := buildKey[T]()

	svc, err := locator.GetService(key)

	if err != nil {
		panic(err)
	}

	return svc.(T)
}
