package graphstore

import (
	"github.com/jbenet/goprocess"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type listenerSlot struct {
	listener IndexedGraphListener
	queue    chan psi.Node
	proc     goprocess.Process
}

type IndexedGraphListener interface {
	OnNodeUpdated(node psi.Node)
}

type IndexedGraphListenerFunc func(node psi.Node)

func (f IndexedGraphListenerFunc) OnNodeUpdated(node psi.Node) {
	f(node)
}
