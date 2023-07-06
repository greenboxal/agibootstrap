package platform

type EventLoop interface {
	Dispatch(f func())

	EnterNestedEventLoop(wait bool)
	ExitNestedEventLoop()
}

type Platform interface {
	EventLoop() EventLoop
}

var instance Platform

func SetInstance(p Platform) { instance = p }
func Instance() Platform     { return instance }

func Dispatch(f func()) {
	Instance().EventLoop().Dispatch(f)
}
