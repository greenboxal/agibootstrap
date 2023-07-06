package graphstore

import (
	"context"
	"sync"

	"github.com/ipld/go-ipld-prime"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type lazyEdge struct {
	psi.EdgeBase

	mu       sync.Mutex
	resolved bool

	frozen *psi.FrozenEdge
	link   ipld.Link
}

func newLazyEdge(from psi.Node, key psi.EdgeKey, link ipld.Link, frozen *psi.FrozenEdge) *lazyEdge {
	le := &lazyEdge{}

	le.SetKey(key)
	le.SetFrom(from)

	return le
}

func (le *lazyEdge) To() psi.Node {
	if !le.resolved {
		if err := le.resolve(); err != nil {
			panic(err)
		}
	}

	return le.EdgeBase.To()
}

func (le *lazyEdge) resolve() (err error) {
	var to psi.Node

	le.mu.Lock()
	defer le.mu.Unlock()

	g := le.Graph().(*IndexedGraph)
	fe := le.frozen
	ctx := context.Background()

	if (fe.Key.Kind == psi.EdgeKindChild || fe.ToPath == nil) && fe.ToLink != nil {
		frozen, err := g.store.GetNodeByCid(ctx, fe.ToLink)

		if err != nil {
			return err
		}

		to, err = g.LoadNode(ctx, frozen)
	} else if fe.ToPath != nil {
		to, err = g.ResolveNode(ctx, *fe.ToPath)

		if err != nil {
			return err
		}
	}

	le.SetTo(to)

	return
}
