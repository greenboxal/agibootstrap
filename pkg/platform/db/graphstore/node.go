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

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type cachedNode struct {
	mu sync.Mutex

	g      *IndexedGraph
	parent *cachedNode

	path   psi.Path
	frozen *psi.FrozenNode
	edges  []*psi.FrozenEdge
	link   ipld.Link
	node   psi.Node

	lastFenceId uint64
}

func (c *cachedNode) FrozenNode() *psi.FrozenNode { return c.frozen }
func (c *cachedNode) Node() psi.Node              { return c.node }
func (c *cachedNode) CommittedVersion() int64     { return c.frozen.Version }
func (c *cachedNode) CommittedLink() ipld.Link    { return c.link }
func (c *cachedNode) LastFenceID() uint64         { return c.lastFenceId }

func (c *cachedNode) Load(ctx context.Context) error {
	if err := c.Preload(ctx); err != nil {
		return err
	}

	if err := c.Refresh(ctx); err != nil {
		return err
	}

	return nil
}

func (c *cachedNode) Preload(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.node != nil {
		return nil
	}

	if c.frozen == nil {
		if c.link == nil {
			frozen, link, err := c.g.store.GetNodeByPath(ctx, c.g.root.UUID(), c.path)

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

	typ := psi.NodeTypeByName(c.frozen.Type)

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

	return nil
}

func (c *cachedNode) Refresh(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.node == nil {
		return nil
	}

	if c.frozen == nil {
		return nil
	}

	if len(c.path.Parent().Components()) > 1 {
		c.parent = c.g.getCacheEntry(c.path.Parent(), true)

		if err := c.parent.Preload(ctx); err != nil {
			return err
		}

		c.node.SetParent(c.parent.node)
	}

	edges, err := c.g.store.ListNodeEdges(ctx, c.g.root.UUID(), c.path)

	if err != nil {
		return err
	}

	for edges.Next() {
		fe := edges.Value()

		if fe.Key.Kind != psi.EdgeKindChild {
			edge := newLazyEdge(c.node, fe.Key, nil, fe)

			c.node.UpsertEdge(edge)
		} else if fe.ToLink != nil {
			frozen, err := c.g.store.GetNodeByCid(ctx, fe.ToLink)

			if err != nil {
				return err
			}

			to, err := c.g.LoadNode(ctx, frozen)

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

	psi.UpdateNodeSnapshot(c.node, c)

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

	return c.updateIndex(ctx, batch)
}

func (c *cachedNode) updateFromMemory(ctx context.Context, node psi.Node) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.node == node {
		return nil
	}

	c.node = node
	c.frozen = nil
	c.edges = nil
	c.link = nil

	return nil
}

func (c *cachedNode) updateFromFreezer(ctx context.Context, link ipld.Link, fn *psi.FrozenNode, edges []*psi.FrozenEdge) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.frozen == fn && c.link == link {
		return nil
	}

	if fn == nil {
		f, err := c.g.store.GetNodeByCid(ctx, link)

		if err != nil {
			return err
		}

		fn = f
	}

	c.frozen = fn
	c.link = link
	c.edges = edges

	return c.Load(ctx)
}

func (c *cachedNode) updateIndex(ctx context.Context, batch datastore.Batch) error {
	var err error
	var shouldCommit bool

	if c.node == nil || c.frozen == nil || c.link == nil {
		return nil
	}

	nodeRoot := c.node

	for parent := nodeRoot; parent != nil; parent = parent.Parent() {
		nodeRoot = parent
	}

	if nodeRoot, ok := nodeRoot.(psi.UniqueNode); ok {
		root := nodeRoot.UUID()

		if batch == nil {
			batch, err = c.g.store.ds.Batch(ctx)

			if err != nil {
				return err
			}

			shouldCommit = true
		}

		if err := c.g.store.IndexNode(ctx, batch, root, c.frozen, c.link); err != nil {
			return err
		}

		for i, edge := range c.edges {
			link := c.frozen.Edges[i]

			if err := c.g.store.IndexEdge(ctx, batch, root, edge, link); err != nil {
				return err
			}
		}
	}

	if shouldCommit {
		if err := batch.Commit(ctx); err != nil {
			return err
		}
	}

	return nil
}
