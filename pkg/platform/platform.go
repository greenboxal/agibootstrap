package platform

import "github.com/greenboxal/agibootstrap/pkg/platform/inject"

type EventLoop interface {
	Dispatch(f func())

	EnterNestedEventLoop(wait bool)
	ExitNestedEventLoop()
}

type Platform interface {
	EventLoop() EventLoop
	ServiceProvider() inject.ServiceProvider
}

var instance Platform

func SetInstance(p Platform)                  { instance = p }
func Instance() Platform                      { return instance }
func ServiceProvider() inject.ServiceProvider { return instance.ServiceProvider() }

func Dispatch(f func()) {
	Instance().EventLoop().Dispatch(f)
}
