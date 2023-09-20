package inject

import (
	"context"
	"sync"

	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"
)

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

func WithServiceRegistry(sr *ServiceRegistry) ServiceProviderOption {
	return func(s *serviceProvider) {
		sr.ApplyTo(s)
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

func (sp *serviceProvider) GetRegistration(key ServiceKey, create bool) (ServiceRegistration, error) {
	return sp.getRegistration(key, create)
}
func (sp *serviceProvider) getRegistration(key ServiceKey, create bool) (ServiceRegistration, error) {
	getRLocked := func() *serviceRegistration {
		sp.mu.RLock()
		defer sp.mu.RUnlock()

		return sp.registrations[key]
	}

	if r := getRLocked(); r != nil {
		return r, nil
	}

	if !create {
		return nil, errors.Wrap(ErrServiceNotFound, 0)
	}

	def := sp.getDefinition(key)

	if def == nil {
		if sp.parent != nil {
			return sp.parent.GetRegistration(key, create)
		}

		return nil, errors.Wrap(ErrServiceDefinitionNotFound, 0)
	}

	sp.mu.Lock()
	defer sp.mu.Unlock()

	if r := sp.registrations[key]; r != nil {
		return r, nil
	}

	r := &serviceRegistration{
		sp:  sp,
		key: key,
		def: def,
	}

	sp.registrations[key] = r

	return r, nil
}

func (sp *serviceProvider) GetService(key ServiceKey) (any, error) {
	reg, err := sp.getRegistration(key, true)

	if err != nil {
		return nil, err
	}

	if reg == nil {
		if sp.parent != nil {
			return sp.parent.GetService(key)
		}

		return nil, ErrServiceNotFound
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

			if len(sp.shutdownHooks) == 0 {
				done = true
				sp.mu.Unlock()
				return
			}

			hook := sp.shutdownHooks[0]
			sp.shutdownHooks = sp.shutdownHooks[1:]
			sp.mu.Unlock()

			if err := hook(ctx); err != nil {
				merr = multierror.Append(merr, err)
			}
		}()
	}

	return merr
}

func (sp *serviceProvider) AppendShutdownHook(f func(ctx context.Context) error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	sp.shutdownHooks = append(sp.shutdownHooks, f)
}

func (sp *serviceProvider) notifyClosed(key ServiceKey) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	delete(sp.registrations, key)
}
