package graphstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type cachedNode struct {
	mu sync.Mutex

	g *IndexedGraph

	id   int64
	path psi.Path

	frozen *psi.FrozenNode
	edges  []*psi.FrozenEdge
	link   ipld.Link
	node   psi.Node

	loaded      bool
	initialized bool

	lastFenceId uint64
}

func (c *cachedNode) ID() int64                   { return c.id }
func (c *cachedNode) Path() psi.Path              { return c.path }
func (c *cachedNode) FrozenNode() *psi.FrozenNode { return c.frozen }
func (c *cachedNode) Node() psi.Node              { return c.node }
func (c *cachedNode) CommitVersion() int64        { return c.frozen.Version }
func (c *cachedNode) CommitLink() ipld.Link       { return c.link }
func (c *cachedNode) LastFenceID() uint64         { return c.lastFenceId }

func (c *cachedNode) Load(ctx context.Context) error {
	if err := c.Preload(ctx); err != nil {
		return err
	}

	if c.node == nil {
		return psi.ErrNodeNotFound
	}

	if !c.initialized {
		if c.path.Len() > 0 {
			parentEntry := c.g.getCacheEntry(c.path.Parent(), true)

			if parentEntry.node == nil {
				if err := parentEntry.Load(ctx); err != nil {
					panic(err)
				}
			}
		}

		if err := c.Refresh(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (c *cachedNode) Preload(ctx context.Context) error {
	defer c.update(ctx)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.loaded {
		return nil
	}

	if c.frozen == nil {
		if c.link == nil {
			frozen, link, err := c.g.store.GetNodeByPath(ctx, c.path)

			if err != nil {
				return err
			}

			c.frozen = frozen
			c.link = link
		} else {
			frozen, err := c.g.store.GetNodeByCid(ctx, c.link)

			if err != nil {
				return err
			}

			c.frozen = frozen
		}
	}

	if c.frozen == nil {
		return psi.ErrNodeNotFound
	}

	typ := psi.NodeTypeByName(c.frozen.Type)

	if typ == nil {
		return fmt.Errorf("unknown node type %q", c.frozen.Type)
	}

	if c.frozen.Data != nil && !typ.Definition().IsRuntimeOnly {
		rawNode, err := c.g.store.lsys.Load(
			ipld.LinkContext{Ctx: ctx},
			c.frozen.Data,
			typ.Type().IpldPrototype(),
		)

		if err != nil {
			return err
		}

		n, ok := typesystem.TryUnwrap[psi.Node](rawNode)

		if !ok {
			return fmt.Errorf("expected node, got %T", rawNode)
		}

		c.node = n
	}

	if c.node != nil {
		c.loaded = true
	}

	return nil
}

func (c *cachedNode) Refresh(ctx context.Context) error {
	defer c.update(ctx)

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.loaded {
		return nil
	}

	if c.frozen == nil {
		return psi.ErrNodeNotFound
	}

	edges, err := c.g.store.ListNodeEdges(ctx, c.path)

	if err != nil {
		return err
	}

	for edges.Next() {
		fe := edges.Value()

		if fe.Key.Kind != psi.EdgeKindChild {
			edge := psi.NewLazyEdge(c.g, fe.Key, c.node, c.resolveEdge)

			c.node.UpsertEdge(edge)
		} else if fe.ToLink != nil {
			frozen, err := c.g.store.GetNodeByCid(ctx, fe.ToLink)

			if err != nil {
				return err
			}

			to, err := c.g.LoadNode(ctx, frozen)

			if err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
				c.g.logger.Warn(err)
				return err
			}

			if to == nil {
				continue
			}

			idx := fe.Key.Index

			if idx >= int64(len(c.node.Children())) {
				idx = int64(len(c.node.Children()))
			}

			c.node.InsertChildrenAt(int(fe.Key.Index), to)
		}
	}

	if c.node.PsiNode() == nil {
		typ := psi.NodeTypeByName(c.frozen.Type)

		typ.InitializeNode(c.node)
	}

	c.initialized = true

	return nil
}

func (c *cachedNode) Commit(ctx context.Context, batch datastore.Batch) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.node == nil {
		return nil
	}

	fn, edges, link, err := c.g.store.FreezeNode(ctx, c.node)

	if err != nil {
		panic(err)
	}

	if c.link != nil && link.Binary() == c.link.Binary() {
		return nil
	}

	records := make([]WalRecord, 2+len(edges))

	records[0] = BuildWalRecord(WalOpUpdateNode, link.(cidlink.Link).Cid)

	for i, edgeLink := range fn.Edges {
		records[i+1] = BuildWalRecord(WalOpUpdateEdge, edgeLink.Cid)
	}

	records[len(records)-1] = BuildWalRecord(WalOpFence, cid.Undef)

	rid, err := c.g.wal.WriteRecords(records...)

	if err != nil {
		return err
	}

	c.frozen = fn
	c.link = link
	c.edges = edges
	c.lastFenceId = rid

	c.g.logger.Debugw("Commit node", "path", c.node.CanonicalPath().String(), "link", link.String())

	if err := c.commitIndex(ctx, batch); err != nil {
		return err
	}

	c.g.nodeUpdateQueue <- nodeUpdateRequest{
		Fence: rid,
		Node:  c.node,
	}

	return nil
}

func (c *cachedNode) update(ctx context.Context) {
	if c.node == nil || c.node.PsiNode() == nil {
		return
	}

	if c.id == -1 {
		if c.frozen != nil {
			c.id = c.frozen.Index
		}

		c.id = int64(c.g.bmp.Allocate())
	}

	if c.id != -1 && c.node != nil {
		c.node.PsiNodeBase().SetSnapshot(c)
	}
}

func (c *cachedNode) updateNode(node psi.Node) error {
	defer c.update(context.Background())

	if c.node == node {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.node == node {
		return nil
	}

	c.node = node

	c.invalidate()

	return nil
}

func (c *cachedNode) commitIndex(ctx context.Context, batch datastore.Batch) error {
	var err error
	var shouldCommit bool

	if c.node == nil || c.frozen == nil || c.link == nil {
		return nil
	}

	if batch == nil {
		batch, err = c.g.store.ds.Batch(ctx)

		if err != nil {
			return err
		}

		shouldCommit = true
	}

	if err := c.g.store.IndexNode(ctx, batch, c.frozen, c.link); err != nil {
		return err
	}

	for i, edge := range c.edges {
		link := c.frozen.Edges[i]

		if err := c.g.store.IndexEdge(ctx, batch, edge, link); err != nil {
			return err
		}
	}

	if shouldCommit {
		if err := batch.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (c *cachedNode) Remove(ctx context.Context, n psi.Node) error {
	if n.CanonicalPath().IsRelative() {
		return fmt.Errorf("node path must be absolute")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.node = nil
	c.frozen = nil
	c.edges = nil
	c.link = nil

	batch, err := c.g.store.ds.Batch(ctx)

	if err != nil {
		return err
	}

	for _, edge := range c.edges {
		if err := c.g.store.RemoveEdgeFromIndex(ctx, batch, c.path, edge.Key); err != nil {
			return err
		}
	}

	if err := c.g.store.RemoveNodeFromIndex(ctx, batch, c.path); err != nil {
		return err
	}

	if err := batch.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (c *cachedNode) resolveEdge(ctx context.Context, g psi.Graph, from psi.Node, key psi.EdgeKey) (psi.Node, error) {
	for _, e := range c.edges {
		if e.Key != key {
			continue
		}

		if n, err := c.g.ResolveEdge(ctx, e); err == nil {
			return n, nil
		}

		break
	}

	return nil, psi.ErrNodeNotFound
}
func (c *cachedNode) invalidate() {
	c.loaded = false
	c.initialized = false
}
