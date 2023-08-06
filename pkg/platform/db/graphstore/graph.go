package graphstore

import (
	"context"
	"os"
	"path"
	"sync"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	"github.com/jbenet/goprocess"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	graphfs2 "github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
)

var logger = logging.GetLogger("graphstore")

type nodeUpdateRequest struct {
	Fence  uint64
	Node   psi.Node
	Frozen *psi.FrozenNode
	Edges  []*psi.FrozenEdge
	Link   ipld.Link
}

type IndexedGraph struct {
	logger *zap.SugaredLogger
	mu     sync.RWMutex

	root psi.UniqueNode

	ds         datastore.Batching
	journal    *graphfs2.Journal
	checkpoint graphfs2.Checkpoint

	vg *graphfs2.VirtualGraph
	lg *online.LiveGraph

	proc            goprocess.Process
	nodeUpdateQueue chan nodeUpdateRequest
	closeCh         chan struct{}

	listeners []*listenerSlot
}

func NewIndexedGraph(ds datastore.Batching, walPath string, root psi.UniqueNode) (*IndexedGraph, error) {
	if err := os.MkdirAll(walPath, 0755); err != nil {
		return nil, err
	}

	journal, err := graphfs2.OpenJournal(walPath)

	if err != nil {
		return nil, err
	}

	checkpoint, err := graphfs2.OpenFileCheckpoint(path.Join(walPath, "ckpt"))

	if err != nil {
		return nil, err
	}

	g := &IndexedGraph{
		logger: logging.GetLogger("graphstore"),

		root: root,
		ds:   ds,

		journal:    journal,
		checkpoint: checkpoint,

		closeCh:         make(chan struct{}),
		nodeUpdateQueue: make(chan nodeUpdateRequest, 8192),
	}

	return g, nil
}

func (g *IndexedGraph) Root() psi.UniqueNode            { return g.root }
func (g *IndexedGraph) Store() *Store                   { return nil }
func (g *IndexedGraph) LinkSystem() *linking.LinkSystem { return nil }
func (g *IndexedGraph) DataStore() datastore.Batching   { return g.ds }

func (g *IndexedGraph) AddListener(l IndexedGraphListener) {
	g.mu.Lock()
	defer g.mu.Unlock()

	index := slices.IndexFunc(g.listeners, func(s *listenerSlot) bool {
		return s.listener == l
	})

	if index != -1 {
		return
	}

	s := &listenerSlot{
		g:        g,
		listener: l,
		queue:    make(chan psi.Node, 128),
	}

	s.proc = goprocess.SpawnChild(g.proc, s.run)

	g.listeners = append(g.listeners, s)
}

func (g *IndexedGraph) RemoveListener(l IndexedGraphListener) {
	g.mu.Lock()
	defer g.mu.Unlock()

	index := slices.IndexFunc(g.listeners, func(s *listenerSlot) bool {
		return s.listener == l
	})

	if index == -1 {
		return
	}

	s := g.listeners[index]

	g.listeners = slices.Delete(g.listeners, index, index+1)

	if err := s.proc.Close(); err != nil {
		panic(err)
	}
}

func (g *IndexedGraph) Shutdown(ctx context.Context) error {
	if g.closeCh != nil {
		close(g.closeCh)
		g.closeCh = nil
	}

	if err := g.proc.Close(); err != nil {
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.journal != nil {
		if err := g.journal.Close(); err != nil {
			return err
		}

		g.journal = nil
	}

	if g.checkpoint != nil {
		if err := g.checkpoint.Close(); err != nil {
			return err
		}

		g.checkpoint = nil
	}

	if g.vg != nil {
		if err := g.vg.Close(ctx); err != nil {
			return err
		}

		g.vg = nil
	}

	return nil
}

func (g *IndexedGraph) notifyNodeUpdated(ctx context.Context, node psi.Node) {
	g.dispatchListeners(node)
}

func (g *IndexedGraph) dispatchListeners(node psi.Node) {
	if node == nil {
		return
	}

	for _, l := range g.listeners {
		if l.queue != nil {
			l.queue <- node
		}
	}
}

func (g *IndexedGraph) LiveGraph() *online.LiveGraph { return g.lg }
