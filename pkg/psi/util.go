package psi

type InvalidationListener interface {
	OnInvalidated(n Node)
}

type invalidationListenerFunc struct{ f func(n Node) }

func InvalidationListenerFunc(f func(n Node)) InvalidationListener {
	return &invalidationListenerFunc{f: f}
}

func (f *invalidationListenerFunc) OnInvalidated(n Node) { f.f(n) }
