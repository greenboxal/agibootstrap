package inject

import "context"

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

func (rc *resolutionContext) GetRegistration(key ServiceKey, create bool) (ServiceRegistration, error) {
	return rc.sp.getRegistration(key, create)
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
