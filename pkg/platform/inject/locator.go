package inject

type ServiceLocator interface {
	GetService(key ServiceKey) (any, error)
	GetRegistration(key ServiceKey, create bool) (ServiceRegistration, error)
}
