package inject

import (
	"context"
	"io"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

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

func (sr *serviceRegistration) GetKey() ServiceKey               { return sr.key }
func (sr *serviceRegistration) GetDefinition() ServiceDefinition { return *sr.def }

func (sr *serviceRegistration) GetInstance(ctx ResolutionContext) (any, error) {
	if instance, err := (func() (any, error) {
		sr.mu.RLock()
		defer sr.mu.RUnlock()

		if sr.closed {
			return nil, ErrServiceClosed
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
		return nil, ErrServiceClosed
	}

	if sr.instance != nil {
		return sr.instance, nil
	}

	if len(sr.deps) != len(sr.def.Dependencies) {
		sr.depInstances = make([]any, len(sr.def.Dependencies))
		sr.depInstances = make([]any, len(sr.deps))
	}

	for i, depKey := range sr.def.Dependencies {
		dep, err := sr.sp.getRegistration(depKey, true)

		if err != nil {
			return nil, err
		}

		if dep == nil {
			return nil, errors.Wrapf(ErrServiceNotFound, "dependency %s not found", depKey)
		}

		sr.deps[i] = dep.(*serviceRegistration)
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
		err := shutdown.Stop(ctx)

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
