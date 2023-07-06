package actions

import "github.com/greenboxal/agibootstrap/pkg/platform/project"

type Action interface {
	Run(ctx Context) error
}

type ActionFunc func(ctx Context) error

func (f ActionFunc) Run(ctx Context) error { return f(ctx) }

type Context interface {
	Context() Context
	Project() project.Project
}
