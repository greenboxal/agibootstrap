package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Route struct {
	Name string
	Root Node
}

type Router struct {
	ContainerBase

	Routes       []*Route
	CurrentRoute obsfx.SimpleProperty[*Route]
}

func NewRouter(routes ...*Route) *Router {
	r := &Router{
		Routes: routes,
	}

	obsfx.ObserveChange(&r.CurrentRoute, func(old, new *Route) {
		if old != nil {
			old.Root.SetParent(nil)
		}

		if new != nil {
			new.Root.SetParent(r)
		}
	})

	r.Init(r)

	return r
}

func (r *Router) Init(self psi.Node) {
	r.ContainerBase.Init(self)
}

func (r *Router) Update(ctx context.Context) error {
	return r.ContainerBase.Update(ctx)
}
