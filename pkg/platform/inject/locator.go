package inject

type ServiceLocator interface {
	GetService(key ServiceKey) (any, error)
}
