package online

import (
	"context"
	"sync"
	"time"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/psidb/graphfs"
)

type liveNodeFlag uint32

const (
	liveNodeFlagLoaded liveNodeFlag = 1 << iota
	liveNodeFlagDirty
	liveNodeFlagNew
	liveNodeFlagPrefetched
	liveNodeFlagEdgesPrefetched
	liveNodeFlagHasInstance
	liveNodeFlagEdgesPopulated
	liveNodeFlagConnected

	liveNodeFlagReady              = liveNodeFlagLoaded | liveNodeFlagEdgesPopulated | liveNodeFlagConnected
	liveNodeFlagNone  liveNodeFlag = 0
)

type LiveNode struct {
	mu sync.RWMutex

	g     *LiveGraph
	flags liveNodeFlag

	cachedIndex int64
	cachedPath  *psi.Path

	inode  *graphfs.INode
	dentry *graphfs.CacheEntry
	handle graphfs.NodeHandle

	frozen *graphfs.SerializedNode
	edges  map[psi.EdgeKey]*LiveEdge

	parent *LiveNode
	node   psi.Node
}

func NewLiveNode(g *LiveGraph) *LiveNode {
	ln := &LiveNode{
		g:           g,
		cachedIndex: -1,
		edges:       map[psi.EdgeKey]*LiveEdge{},
	}

	return ln
}

func (ln *LiveNode) ID() int64                      { return ln.inode.ID() }
func (ln *LiveNode) Node() psi.Node                 { return ln.node }
func (ln *LiveNode) Path() psi.Path                 { return ln.dentry.Path() }
func (ln *LiveNode) FrozenNode() *psi.FrozenNode    { panic("deprecated") }
func (ln *LiveNode) FrozenEdges() []*psi.FrozenEdge { panic("deprecated") }
func (ln *LiveNode) CommitLink() ipld.Link          { return nil }
func (ln *LiveNode) CommitVersion() int64 {
	if ln.frozen != nil {
		return ln.frozen.Version
	}

	return 0
}
func (ln *LiveNode) LastFenceID() uint64 { return 0 }

func (ln *LiveNode) Get(ctx context.Context) (psi.Node, error) {
	if ln.node == nil {
		if err := ln.prefetch(ctx); err != nil {
			return nil, err
		}

		err := ln.recreateNode(ctx, ln.node)

		if err != nil {
			return nil, err
		}
	}

	if ln.node == nil {
		return nil, psi.ErrNodeNotFound
	}

	if ln.frozen != nil && ln.node.PsiNode() == nil {
		typ := psi.NodeTypeByName(ln.frozen.Type)

		typ.InitializeNode(ln.node)
	}

	return ln.node, nil
}

func (ln *LiveNode) Load(ctx context.Context) error {
	if ln.flags&liveNodeFlagLoaded != 0 {
		return nil
	}

	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.flags&liveNodeFlagLoaded != 0 {
		return nil
	}

	if err := ln.prefetch(ctx); err != nil && err != psi.ErrNodeNotFound {
		return err
	}

	err := ln.recreateNode(ctx, ln.node)

	if err != nil {
		return err
	}

	ln.flags |= liveNodeFlagLoaded

	return nil
}

func (ln *LiveNode) Save(ctx context.Context) error {
	keepCookie := time.Now().Unix()

	ln.mu.Lock()
	defer ln.mu.Unlock()

	if p := ln.node.Parent(); p != nil {
		if ln.parent.node != p {
			ln.parent = ln.g.nodeForNode(p)
		}
	} else {
		ln.parent = nil
	}

	//if ln.flags&liveNodeFlagDirty == 0 {
	//	return nil
	//}

	if ln.flags&liveNodeFlagPrefetched == 0 {
		if err := ln.prefetchNode(ctx); err != nil && err != psi.ErrNodeNotFound {
			return err
		}
	}

	if ln.frozen == nil {
		ln.frozen = &graphfs.SerializedNode{}
	}

	typ := ln.node.PsiNodeType()

	ln.frozen.Index = ln.cachedIndex
	ln.frozen.Path = ln.node.CanonicalPath()
	ln.frozen.Version = ln.frozen.Version + 1
	ln.frozen.Type = typ.Name()

	if ln.parent != nil {
		ln.frozen.Parent = ln.parent.cachedIndex
	} else {
		ln.frozen.Parent = -1
	}

	if !typ.Definition().IsRuntimeOnly {
		data, err := ipld.Encode(typesystem.Wrap(ln.node), dagjson.Encode)

		if err != nil {
			return err
		}

		ln.frozen.Data = data
		ln.frozen.Flags |= graphfs.NodeFlagHasData
	} else {
		ln.frozen.Flags &= ^graphfs.NodeFlagHasData
	}

	nh, err := ln.reopen(ctx, graphfs.OpenNodeOptions{
		Flags: graphfs.OpenNodeFlagsCreate | graphfs.OpenNodeFlagsWrite | graphfs.OpenNodeFlagsAppend | graphfs.OpenNodeFlagsRead,
	})

	if err != nil {
		return err
	}

	for _, v := range ln.node.Children() {
		k := v.CanonicalPath().Name()
		le, err := ln.prefetchEdge(ctx, k, nil)

		if err != nil {
			return err
		}

		ln.node.UpsertEdge(le.ReplaceTo(v))

		if err := le.Save(ctx, nh); err != nil {
			return err
		}

		le.cookie = keepCookie
	}

	for edges := ln.node.Edges(); edges.Next(); {
		e := edges.Value()
		le, err := ln.prefetchEdge(ctx, e.Key().GetKey(), nil)

		if err != nil {
			return err
		}

		if e != le {
			ln.node.UpsertEdge(le.ReplaceTo(e.To()))
		}

		if err := le.Save(ctx, nh); err != nil {
			return err
		}

		le.cookie = keepCookie
	}

	for k, v := range ln.edges {
		if v.cookie != keepCookie {
			delete(ln.edges, k)
		}
	}

	if _, err := nh.Write(ctx, ln.frozen); err != nil {
		return err
	}

	ln.flags &= ^liveNodeFlagDirty
	ln.flags &= ^liveNodeFlagNew

	return nil
}

func (ln *LiveNode) recreateNode(ctx context.Context, n psi.Node) error {
	if ln.frozen == nil && n != nil {
		ln.frozen = &graphfs.SerializedNode{}
	}

	if n == nil {
		if ln.frozen.Flags&graphfs.NodeFlagHasData != 0 {
			typ := psi.NodeTypeByName(ln.frozen.Type)

			if typ == nil {
				panic("unknown node type")
			}

			wrapped, err := ipld.DecodeUsingPrototype(ln.frozen.Data, dagjson.Decode, typ.Type().IpldPrototype())

			if err != nil {
				return err
			}

			n, _ = typesystem.TryUnwrap[psi.Node](wrapped)
		} else if ln.frozen.Flags&graphfs.NodeFlagHasDataLink != 0 {
			_, err := cid.Cast(ln.frozen.Data)

			if err != nil {
				return err
			}

			panic("TODO: implement")
		}
	}

	if n == nil {
		return psi.ErrNodeNotFound
	}

	ln.updateNode(n)

	return ln.populateNode(ctx)
}

func (ln *LiveNode) populateNode(ctx context.Context) error {
	if ln.flags&liveNodeFlagEdgesPopulated == 0 {
		for _, edge := range ln.edges {
			existing := ln.node.GetEdge(edge.key)

			if existing != nil && existing != edge {
				continue
			}

			ln.node.UpsertEdge(edge)
		}

		ln.flags |= liveNodeFlagEdgesPopulated
	}

	if ln.flags&liveNodeFlagConnected == 0 {
		if ln.parent != nil {
			parent, err := ln.parent.Get(ctx)

			if err != nil {
				return nil
			}

			idx := int(ln.dentry.Name().Index)

			if idx < 0 || idx >= parent.ChildrenList().Len() {
				idx = parent.ChildrenList().Len()
			}

			parent.InsertChildrenAt(idx, ln.node)
		}

		ln.flags |= liveNodeFlagConnected
	}

	return nil
}

func (ln *LiveNode) prefetch(ctx context.Context) error {
	if err := ln.prefetchNode(ctx); err != nil {
		return err
	}

	if err := ln.prefetchEdges(ctx); err != nil {
		return err
	}

	return nil
}

func (ln *LiveNode) prefetchNode(ctx context.Context) error {
	if ln.flags&liveNodeFlagPrefetched != 0 {
		return nil
	}

	nh, err := ln.reopen(context.Background(), graphfs.OpenNodeOptions{
		Flags: graphfs.OpenNodeFlagsRead,
	})

	if err == psi.ErrNodeNotFound {
		ln.flags |= liveNodeFlagPrefetched | liveNodeFlagNew
		return nil
	} else if err != nil {
		return nil
	}

	sn, err := nh.Read(ctx)

	if err != nil {
		return nil
	}

	ln.frozen = sn
	ln.flags |= liveNodeFlagPrefetched

	return nil
}

func (ln *LiveNode) prefetchEdges(ctx context.Context) error {
	if ln.flags&liveNodeFlagEdgesPrefetched != 0 {
		return nil
	}

	if ln.flags&liveNodeFlagNew != 0 {
		ln.flags |= liveNodeFlagEdgesPrefetched
		return nil
	}

	edges, err := ln.handle.ReadEdges(ctx)

	if err == psi.ErrNodeNotFound {
		ln.flags |= liveNodeFlagEdgesPrefetched
		return nil
	} else if err != nil {
		return err
	}

	for edges.Next() {
		e := edges.Value()

		_, err := ln.prefetchEdge(ctx, e.Key, e)

		if err != nil {
			return err
		}
	}

	ln.flags |= liveNodeFlagEdgesPrefetched

	return nil
}

func (ln *LiveNode) prefetchEdge(ctx context.Context, key psi.EdgeKey, frozen *graphfs.SerializedEdge) (le *LiveEdge, err error) {
	if e := ln.edges[key]; e != nil {
		if frozen == nil {
			f, err := ln.handle.ReadEdge(ctx, key)

			if err != nil && err != psi.ErrNodeNotFound {
				return nil, err
			}

			frozen = f
		}

		if frozen != nil {
			e.update(frozen)
		}

		return e, nil
	}

	dentry, err := ln.dentry.Lookup(ctx, key.AsPathElement())

	if err != nil {
		return nil, err
	}

	e := NewLiveEdge(ln, dentry)

	ln.edges[key] = e

	if frozen == nil {
		f, err := ln.handle.ReadEdge(ctx, key)

		if err != nil && err != psi.ErrNodeNotFound {
			return nil, err
		}

		frozen = f
	}

	if frozen != nil {
		e.update(frozen)
	}

	return e, nil
}

func (ln *LiveNode) reopen(ctx context.Context, opts graphfs.OpenNodeOptions) (graphfs.NodeHandle, error) {
	if ln.parent != nil {
		if _, err := ln.parent.reopen(ctx, opts); err != nil {
			return nil, err
		}
	}

	if ln.node != nil && (ln.dentry == nil || !ln.dentry.Path().Equals(ln.node.CanonicalPath())) {
		ce, err := ln.g.vg.Resolve(ctx, ln.node.CanonicalPath())

		if err != nil {
			return nil, err
		}

		ln.updateDentry(ce)
	}

	if h := ln.handle; h != nil && h.Options() == opts {
		return h, nil
	}

	nh, err := ln.dentry.INodeOperations().Create(ctx, ln.dentry, opts)

	if err != nil {
		return nil, err
	}

	ln.updateOpenHandle(nh)

	return ln.handle, nil
}

func (ln *LiveNode) updateDentry(dentry *graphfs.CacheEntry) {
	if ln.dentry == dentry {
		return
	}

	if ln.dentry != nil {
		ln.dentry.Unref()
		ln.dentry = nil
	}

	if dentry != nil {
		ln.dentry = dentry.Get()

		ln.updateInode(ln.dentry.Inode())
	} else {
		ln.updateInode(nil)
	}

	ln.g.updateNodeCache(ln)
}

func (ln *LiveNode) updateInode(ino *graphfs.INode) {
	if ln.inode == ino {
		return
	}

	if ln.inode != nil {
		ln.inode.Unref()
		ln.inode = nil
	}

	if ino != nil {
		ln.inode = ino.Get()

		if ln.handle != nil && ln.handle.Inode() != ino {
			panic("unexpected")
		}
	}

	ln.g.updateNodeCache(ln)
}

func (ln *LiveNode) updateOpenHandle(nh graphfs.NodeHandle) {
	if ln.handle != nil {
		if err := ln.handle.Close(); err != nil {
			logger.Error(err)
		}

		ln.handle = nil
	}

	ln.handle = nh

	if ln.handle != nil {
		ln.updateInode(ln.handle.Inode())
	}
}

func (ln *LiveNode) updateNode(node psi.Node) {
	if ln.node == node {
		return
	}

	if ln.node != nil {
		ln.node.PsiNodeBase().SetSnapshot(nil)
	}

	ln.node = node

	if ln.node != nil {
		ln.node.PsiNodeBase().SetSnapshot(ln)

		if p := ln.node.Parent(); p != nil {
			ln.parent = ln.g.nodeForNode(p)
		} else {
			ln.parent = nil
		}

		ln.flags |= liveNodeFlagHasInstance
	} else {
		ln.parent = nil

		ln.flags &= ^liveNodeFlagHasInstance
	}

	ln.g.updateNodeCache(ln)
}

func (ln *LiveNode) Close() error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.handle != nil {
		if err := ln.handle.Close(); err != nil {
			return err
		}
	}

	if ln.inode != nil {
		ln.inode.Unref()
		ln.inode = nil
	}

	if ln.dentry != nil {
		ln.dentry.Unref()
		ln.dentry = nil
	}

	ln.node = nil

	return nil
}
