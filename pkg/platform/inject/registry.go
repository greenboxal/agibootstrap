package inject

import (
	`go.uber.org/fx`
)

type ServiceRegistrationScope int

const (
	ServiceRegistrationScopeSingleton ServiceRegistrationScope = iota
	ServiceRegistrationScopeSession
	ServiceRegistrationScopeTransaction
)

type ScopedServiceDefinition struct {
	ServiceDefinition

	Scope ServiceRegistrationScope
}

type ServiceRegistry struct {
	definitions []ScopedServiceDefinition
}

func (sr *ServiceRegistry) Build(opts ...ServiceProviderOption) ServiceProvider {
	sp := NewServiceProvider(opts...)
	sr.ApplyTo(sp)
	return sp
}

func (sr *ServiceRegistry) ApplyTo(sp ServiceProvider) {
	for _, def := range sr.definitions {
		sp.RegisterService(def.ServiceDefinition)
	}
}

func (sr *ServiceRegistry) Add(def ScopedServiceDefinition) {
	sr.definitions = append(sr.definitions, def)
}

func ProvideRegisteredService[T any](scope ServiceRegistrationScope, factory func(ctx ResolutionContext) (T, error)) fx.Option {
	return fx.Invoke(func(srm *ServiceRegistrationManager, svc T) {
		def := ScopedServiceDefinition{
			ServiceDefinition: Provide(factory),
			Scope:             scope,
		}

		srm.RegisterService(def)
	})
}

func WithRegisteredService[T any](scope ServiceRegistrationScope) fx.Option {
	return fx.Invoke(func(srm *ServiceRegistrationManager, svc T) {
		def := ScopedServiceDefinition{
			ServiceDefinition: ProvideInstance(svc),
			Scope:             scope,
		}

		srm.RegisterService(def)
	})
}

type ServiceRegistrationManager struct {
	Global      ServiceRegistry
	Session     ServiceRegistry
	Transaction ServiceRegistry
}

func NewServiceRegistrationManager() *ServiceRegistrationManager {
	return &ServiceRegistrationManager{}
}

func (srm *ServiceRegistrationManager) GetRegistry(scope ServiceRegistrationScope) *ServiceRegistry {
	switch scope {
	case ServiceRegistrationScopeSingleton:
		return &srm.Global
	case ServiceRegistrationScopeSession:
		return &srm.Session
	case ServiceRegistrationScopeTransaction:
		return &srm.Transaction
	default:
		panic("invalid scope")
	}
}

func (srm *ServiceRegistrationManager) RegisterService(def ScopedServiceDefinition) {
	reg := srm.GetRegistry(def.Scope)
	reg.Add(def)
}
