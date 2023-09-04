package inject

import (
	"context"
	"io"
	"reflect"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

var ServiceNotFound = errors.New("service not found")

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

type serviceProvider struct {
	mu sync.RWMutex

	parent ServiceLocator

	definitions   map[ServiceKey]*ServiceDefinition
	registrations map[ServiceKey]*serviceRegistration

	shutdownHooks []func(ctx context.Context) error
}

type ServiceProviderOption func(*serviceProvider)

func WithParentServiceProvider(sp ServiceLocator) ServiceProviderOption {
	return func(s *serviceProvider) {
		s.parent = sp
	}
}

func NewServiceProvider(options ...ServiceProviderOption) ServiceProvider {
	sp := &serviceProvider{
		definitions:   make(map[ServiceKey]*ServiceDefinition),
		registrations: make(map[ServiceKey]*serviceRegistration),
	}

	for _, option := range options {
		option(sp)
	}

	return sp
}

func (sp *serviceProvider) getDefinition(key ServiceKey) *ServiceDefinition {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	return sp.definitions[key]
}

func (sp *serviceProvider) getRegistration(key ServiceKey, create bool) *serviceRegistration {
	getRLocked := func() *serviceRegistration {
		sp.mu.RLock()
		defer sp.mu.RUnlock()

		return sp.registrations[key]
	}

	if r := getRLocked(); r != nil {
		return r
	}

	if !create {
		return nil
	}

	def := sp.getDefinition(key)

	if def == nil {
		return nil
	}

	sp.mu.Lock()
	defer sp.mu.Unlock()

	if r := sp.registrations[key]; r != nil {
		return r
	}

	r := &serviceRegistration{
		sp:  sp,
		key: key,
		def: def,
	}

	sp.registrations[key] = r

	return r
}

func (sp *serviceProvider) GetService(key ServiceKey) (any, error) {
	reg := sp.getRegistration(key, true)

	if reg == nil {
		if sp.parent != nil {
			return sp.parent.GetService(key)
		}

		return nil, ServiceNotFound
	}

	return reg.GetInstance(rootResolutionContext{serviceProvider: sp})
}

func (sp *serviceProvider) RegisterService(registration ServiceDefinition) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.definitions[registration.Key] = &registration
}

func (sp *serviceProvider) Close(ctx context.Context) error {
	var merr error

	for done := false; !done; {
		func() {
			sp.mu.Lock()
			defer sp.mu.Unlock()

			if len(sp.shutdownHooks) == 0 {
				done = true
				return
			}

			hook := sp.shutdownHooks[0]
			sp.shutdownHooks = sp.shutdownHooks[1:]

			if err := hook(ctx); err != nil {
				merr = multierror.Append(merr, err)
			}
		}()
	}

	return merr
}

func (sp *serviceProvider) AppendShutdownHook(f func(ctx context.Context) error) {
	sp.shutdownHooks = append(sp.shutdownHooks, f)
}

func (sp *serviceProvider) notifyClosed(key ServiceKey) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	delete(sp.registrations, key)
}

type serviceRegistration struct {
	mu sync.RWMutex

	sp  *serviceProvider
	key ServiceKey
	def *ServiceDefinition

	deps         []*serviceRegistration
	depInstances []any

	instance any
	closed   bool
}

type rootResolutionContext struct {
	*serviceProvider
}

func (rc rootResolutionContext) Path() []ServiceKey { return nil }
func (rc rootResolutionContext) AppendShutdownHook(f func(ctx context.Context) error) {
	rc.AppendShutdownHook(f)
}

type resolutionContext struct {
	sp   *serviceProvider
	sr   *serviceRegistration
	path []ServiceKey
}

func (rc *resolutionContext) AppendShutdownHook(f func(ctx context.Context) error) {
	rc.sp.AppendShutdownHook(f)
}

func (rc *resolutionContext) Path() []ServiceKey { return rc.path }

func (rc *resolutionContext) GetService(key ServiceKey) (any, error) {
	for _, dep := range rc.sr.deps {
		if dep.key == key {
			return dep.GetInstance(rc)
		}
	}

	return rc.sp.GetService(key)
}

func (sr *serviceRegistration) GetKey() ServiceKey               { return sr.key }
func (sr *serviceRegistration) GetDefinition() ServiceDefinition { return *sr.def }

func (sr *serviceRegistration) GetInstance(ctx ResolutionContext) (any, error) {
	if instance, err := (func() (any, error) {
		sr.mu.RLock()
		defer sr.mu.RUnlock()

		if sr.closed {
			return nil, errors.New("service closed")
		}

		return sr.instance, nil
	})(); err != nil {
		return nil, err
	} else if instance != nil {
		return instance, nil
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.closed {
		return nil, errors.New("service closed")
	}

	if sr.instance != nil {
		return sr.instance, nil
	}

	if len(sr.deps) != len(sr.def.Dependencies) {
		sr.depInstances = make([]any, len(sr.def.Dependencies))
		sr.depInstances = make([]any, len(sr.deps))
	}

	for i, depKey := range sr.def.Dependencies {
		dep := sr.sp.getRegistration(depKey, true)

		if dep == nil {
			return nil, errors.Errorf("dependency %s not found", depKey)
		}

		sr.deps[i] = dep
	}

	for i, dep := range sr.deps {
		instance, err := dep.GetInstance(ctx)

		if err != nil {
			return nil, err
		}

		sr.depInstances[i] = instance
	}

	if idx := slices.Index(ctx.Path(), sr.key); idx != -1 {
		return nil, errors.Errorf("circular dependency detected: %s", ctx.Path()[idx:])
	}

	instance, err := sr.def.Factory(&resolutionContext{
		sp:   sr.sp,
		sr:   sr,
		path: append(slices.Clone(ctx.Path()), sr.key),
	}, sr.depInstances)

	if err != nil {
		return nil, err
	}

	sr.instance = instance

	sr.sp.AppendShutdownHook(sr.Close)

	return instance, nil
}

func (sr *serviceRegistration) Close(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.closed {
		return nil
	}

	if closer, ok := sr.instance.(io.Closer); ok {
		err := closer.Close()

		if err != nil {
			return err
		}
	} else if shutdown, ok := sr.instance.(ShutdownContext); ok {
		err := shutdown.Shutdown(ctx)

		if err != nil {
			return err
		}
	} else if closer, ok := sr.instance.(CloseContext); ok {
		err := closer.Close(ctx)

		if err != nil {
			return err
		}
	}

	sr.closed = true

	sr.sp.notifyClosed(sr.key)

	return nil
}

type CloseContext interface {
	Close(ctx context.Context) error
}

type ShutdownContext interface {
	Shutdown(ctx context.Context) error
}

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
	return Provide[T](func(ctx ResolutionContext) (T, error) {
		return instance, nil
	})
}

func Register[T any](sp ServiceProvider, factory func(ctx ResolutionContext) (T, error)) {
	sp.RegisterService(Provide[T](factory))
}

func Provide[T any](factory func(ctx ResolutionContext) (T, error)) ServiceDefinition {
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
	return ServiceKey{
		Type: typ,
	}
}
