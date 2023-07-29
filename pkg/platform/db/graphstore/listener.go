package graphstore

import (
	"github.com/jbenet/goprocess"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type IndexedGraphListener interface {
	OnNodeUpdated(node psi.Node)
}

type IndexedGraphListenerFunc func(node psi.Node)

func (f IndexedGraphListenerFunc) OnNodeUpdated(node psi.Node) {
	f(node)
}

type listenerSlot struct {
	g        *IndexedGraph
	queue    chan psi.Node
	proc     goprocess.Process
	listener IndexedGraphListener
}

func (s listenerSlot) run(proc goprocess.Process) {
	for {
		select {
		case <-proc.Closing():
			return

		case n := <-s.queue:
			func() {
				defer func() {
					if r := recover(); r != nil {
						s.g.logger.Error("panic in graph listener", "err", r)
					}
				}()

				s.listener.OnNodeUpdated(n)
			}()
		}
	}
}
