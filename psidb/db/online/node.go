package online

import (
	"context"
	"io"
	"reflect"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	graphfs "github.com/greenboxal/agibootstrap/psidb/db/graphfs"
	"github.com/greenboxal/agibootstrap/psidb/psi"
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
	liveNodeFlagInjected
	liveNodeFlagIsInitializing
	liveNodeFlagInitialized
)

type LiveNode struct {
	mu sync.RWMutex

	g     *LiveGraph
	flags liveNodeFlag

	cachedIndex int64
	cachedPath  *psi.Path
	cachedLink  *cidlink.Link

	inode  *graphfs.INode
	dentry *graphfs.CacheEntry
	handle graphfs.NodeHandle

	frozen *coreapi.SerializedNode
	edges  map[psi.EdgeKey]*LiveEdge

	parent *LiveNode
	node   psi.Node

	path psi.Path
}

func NewLiveNode(g *LiveGraph) *LiveNode {
	ln := &LiveNode{
		g:           g,
		cachedIndex: -1,
		edges:       map[psi.EdgeKey]*LiveEdge{},
	}

	return ln
}

func (ln *LiveNode) ID() int64 {
	if ln.inode == nil {
		return -1
	}

	return ln.inode.ID()
}

func (ln *LiveNode) Node() psi.Node                 { return ln.node }
func (ln *LiveNode) Path() psi.Path                 { return ln.path }
func (ln *LiveNode) FrozenNode() *psi.FrozenNode    { panic("deprecated") }
func (ln *LiveNode) FrozenEdges() []*psi.FrozenEdge { panic("deprecated") }
func (ln *LiveNode) CommitLink() ipld.Link          { return ln.cachedLink }
func (ln *LiveNode) CommitVersion() int64 {
	if ln.frozen != nil {
		return ln.frozen.Version
	}

	return 0
}
func (ln *LiveNode) LastFenceID() uint64 { return 0 }

func (ln *LiveNode) Get(ctx context.Context) (psi.Node, error) {
	if ln.node == nil {
		if err := ln.Load(ctx); err != nil {
			return nil, err
		}
	}

	if ln.node == nil {
		return nil, psi.ErrNodeNotFound
	}

	if ln.flags&liveNodeFlagInitialized == 0 {
		if ln.node.PsiNode() == nil {
			ln.flags |= liveNodeFlagIsInitializing

			typ := ln.g.typeRegistry.NodeTypeByName(ctx, ln.frozen.Type)

			typ.InitializeNode(ln.node)

			ln.node.PsiNodeBase().AttachToGraph(ln.g)

			ln.flags &= ^liveNodeFlagIsInitializing
		}

		if ln.frozen == nil {
			ln.g.markDirty(ln.node)
		}

		err := ln.Node().PsiNodeType().OnAfterNodeLoaded(ctx, ln.node)

		if err != nil {
			return nil, err
		}

		ln.flags |= liveNodeFlagInitialized
	}

	return ln.node, nil
}

func (ln *LiveNode) getUnloaded(ctx context.Context) (psi.Node, error) {
	if ln.node == nil {
		if err := ln.Load(ctx); err != nil {
			return nil, err
		}
	}

	if ln.node == nil {
		return nil, psi.ErrNodeNotFound
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

	if err := ln.prefetch(ctx); err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
		return err
	}

	err := ln.recreateNode(ctx, ln.node)

	if err != nil {
		return err
	}

	ln.flags |= liveNodeFlagLoaded

	return nil
}

func (ln *LiveNode) recreateNode(ctx context.Context, n psi.Node) error {
	if ln.frozen == nil && n != nil {
		ln.frozen = &coreapi.SerializedNode{}
	}

	if n == nil {
		if ln.frozen == nil {
			return psi.ErrNodeNotFound
		}

		if ln.frozen.Flags&coreapi.NodeFlagHasData != 0 {
			typ := ln.g.typeRegistry.NodeTypeByName(ctx, ln.frozen.Type)

			if typ == nil {
				panic("unknown node type")
			}

			wrapped, err := ipld.DecodeUsingPrototype(ln.frozen.Data, dagjson.Decode, typ.Type().IpldPrototype())

			if err != nil {
				return err
			}

			n, _ = typesystem.TryUnwrap[psi.Node](wrapped)
		} else if ln.frozen.Flags&coreapi.NodeFlagHasDataLink != 0 {
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

	if err := ln.populateNode(ctx); err != nil {
		return err
	}

	return nil
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

	if ln.flags&liveNodeFlagInjected == 0 {
		var walkFields func(v reflect.Value) error

		walkFields = func(v reflect.Value) error {
			t := v.Type()

			for i := 0; i < t.NumField(); i++ {
				def := t.Field(i)

				if !def.IsExported() {
					continue
				}

				vf := v.Field(i)

				if def.Anonymous && def.Type.Kind() == reflect.Struct {
					if err := walkFields(vf); err != nil {
						return err
					}
				} else {
					_, ok := def.Tag.Lookup("inject")

					if !ok {
						continue
					}

					key := inject.ServiceKeyFor(def.Type)
					resolved, err := ln.g.sp.GetService(key)

					if err != nil {
						return err
					}

					vf.Set(reflect.ValueOf(resolved))
				}
			}

			return nil
		}

		v := reflect.ValueOf(ln.Node())
		v = reflect.Indirect(v)

		if err := walkFields(v); err != nil {
			return err
		}

		ln.flags |= liveNodeFlagInjected
	}

	if ln.flags&liveNodeFlagConnected == 0 {
		if ln.parent == nil && ln.frozen != nil && ln.Path().Len() > 0 {
			parentPath := ln.frozen.Path.Parent()
			parent, err := ln.g.resolveNodeUnloaded(ctx, parentPath)

			if err != nil {
				return err
			}

			ln.parent = parent
		}

		if ln.parent != nil {
			parent, err := ln.parent.getUnloaded(ctx)

			if err != nil {
				return nil
			}

			/*idx := int(ln.dentry.Name().Index)

			if idx < 0 || idx >= parent.ChildrenList().Len() {
				idx = parent.ChildrenList().Len()
			}

			parent.InsertChildrenAt(idx, ln.node)*/

			ln.node.PsiNodeBase().SetParent(parent)
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

	nh, err := ln.reopen(ctx, graphfs.OpenNodeOptions{
		Flags: graphfs.OpenNodeFlagsRead,
	})

	if errors.Is(err, psi.ErrNodeNotFound) {
		ln.flags |= liveNodeFlagPrefetched | liveNodeFlagNew
		return nil
	} else if err != nil {
		return nil
	}

	sn, err := nh.Read(ctx)

	if err != nil {
		return nil
	}

	link, err := ln.g.lsys.ComputeLink(linkPrototype, typesystem.Wrap(sn))

	if err != nil {
		return err
	}

	clink := link.(cidlink.Link)

	ln.frozen = sn
	ln.cachedLink = &clink
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

	if errors.Is(err, psi.ErrNodeNotFound) {
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

func (ln *LiveNode) prefetchEdge(ctx context.Context, key psi.EdgeKey, frozen *coreapi.SerializedEdge) (le *LiveEdge, err error) {
	if e := ln.edges[key]; e != nil {
		if frozen == nil {
			f, err := ln.handle.ReadEdge(ctx, key)

			if err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
				return nil, err
			}

			frozen = f
		}

		if frozen != nil {
			e.update(frozen)
		}

		return e, nil
	}

	e := NewLiveEdge(ln, key)

	ln.edges[key] = e

	if ln.dentry != nil {
		dentry, err := ln.dentry.Lookup(ctx, key.AsPathElement())

		if err != nil {
			return nil, err
		}

		e.updateDentry(dentry)
	}

	if frozen == nil {
		f, err := ln.handle.ReadEdge(ctx, key)

		if err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
			return nil, err
		}

		frozen = f
	}

	if frozen != nil {
		e.update(frozen)
	}

	return e, nil
}

func (ln *LiveNode) Save(ctx context.Context) error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	//logger.Debugw("Save", "path", ln.path)

	if ln.flags&liveNodeFlagDirty == 0 {
		return nil
	}

	if ln.flags&liveNodeFlagInitialized == 0 {
		ln.flags &= ^liveNodeFlagDirty

		if ln.node != nil {
			ln.g.markClean(ln.node)
		}

		return nil
	}

	if ln.flags&liveNodeFlagPrefetched == 0 {
		if err := ln.prefetchNode(ctx); err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
			return err
		}
	}

	if ln.frozen == nil {
		ln.frozen = &coreapi.SerializedNode{}
	}

	typ := ln.node.PsiNodeType()

	if err := typ.OnBeforeNodeSaved(ctx, ln.node); err != nil {
		return err
	}

	ln.frozen.Type = typ.Name()

	if p := ln.Node().Parent(); p != nil {
		ln.frozen.Path = p.CanonicalPath().Join(ln.node.SelfIdentity())
		ln.frozen.Parent = ln.g.nodeForNode(p).cachedIndex
	} else {
		ln.frozen.Path = ln.node.SelfIdentity()
		ln.frozen.Parent = -1
	}

	if !typ.Definition().IsRuntimeOnly {
		wrapped := typesystem.Wrap(ln.node)
		link, err := ln.g.lsys.ComputeLink(linkPrototype, wrapped)

		if err != nil {
			return err
		}

		clink := link.(cidlink.Link)

		if ln.frozen.Link == nil || !ln.frozen.Link.Equals(clink.Cid) {
			ln.frozen.Version++
		}

		data, err := ipld.Encode(wrapped, dagjson.Encode)

		if err != nil {
			return err
		}

		ln.frozen.Data = data
		ln.frozen.Link = &clink
		ln.frozen.Flags |= coreapi.NodeFlagHasData
	} else {
		ln.frozen.Flags &= ^coreapi.NodeFlagHasData
	}

	nh, err := ln.reopen(ctx, graphfs.OpenNodeOptions{
		Flags: graphfs.OpenNodeFlagsCreate | graphfs.OpenNodeFlagsWrite | graphfs.OpenNodeFlagsAppend | graphfs.OpenNodeFlagsRead,
	})

	if err != nil {
		return err
	}

	ln.frozen.Index = ln.cachedIndex

	if ln.parent != nil {
		ln.frozen.Parent = ln.parent.cachedIndex
	} else {
		ln.frozen.Parent = -1
	}

	for _, v := range ln.node.Children() {
		k := v.SelfIdentity().Name()

		if k.Name == "" && (k.Kind == "" || k.Kind == psi.EdgeKindChild) {
			continue
		}

		le, err := ln.prefetchEdge(ctx, k, nil)

		if err != nil {
			return err
		}

		ln.node.UpsertEdge(le.ReplaceTo(v))
	}

	for edges := ln.node.Edges(); edges.Next(); {
		edge := edges.Value()
		existing := ln.edges[edge.Key().GetKey()]

		if existing == nil {
			le, err := ln.prefetchEdge(ctx, edge.Key().GetKey(), nil)

			if err != nil {
				return err
			}

			ln.markEdgeDirty(le.key)
		}
	}

	for _, le := range ln.edges {
		if le.dentry == nil && ln.dentry != nil {
			ce, err := ln.dentry.Lookup(ctx, le.key.AsPathElement())

			if err != nil {
				return err
			}

			le.updateDentry(ce)
		}

		if err := le.Save(ctx, nh); err != nil {
			return err
		}
	}

	link, err := ln.g.lsys.ComputeLink(linkPrototype, typesystem.Wrap(ln.frozen))

	if err != nil {
		return err
	}

	clink := link.(cidlink.Link)

	if ln.cachedLink == nil || !clink.Equals(ln.cachedLink.Cid) {
		if err := nh.Write(ctx, ln.frozen); err != nil {
			return err
		}

		ln.cachedLink = &clink
	}

	ln.flags &= ^liveNodeFlagDirty
	ln.flags &= ^liveNodeFlagNew

	if ln.node != nil {
		ln.g.markClean(ln.node)
	}

	return nil
}

func (ln *LiveNode) reopen(ctx context.Context, opts graphfs.OpenNodeOptions) (graphfs.NodeHandle, error) {
	if ln.parent != nil {
		if _, err := ln.parent.reopen(ctx, opts); err != nil {
			return nil, err
		}
	}

	if ln.dentry != nil && !ln.dentry.Path().Equals(ln.path) {
		ln.updatePath()

		ce, err := ln.g.graph.Resolve(ctx, ln.path)

		if err != nil {
			return nil, err
		}

		ln.updateDentry(ce)
	}

	if h := ln.handle; h != nil && h.Options() == opts && ln.handle.Entry() == ln.dentry && ln.handle.Inode() == ln.inode {
		return h, nil
	}

	nh, err := ln.g.graph.Open(ctx, ln.path, graphfs.WithOpenNodeOptions(opts))

	if err != nil {
		return nil, err
	}

	ln.updateOpenHandle(nh)

	return ln.handle, nil
}

func (ln *LiveNode) updatePath() {
	if ln.frozen != nil {
		ln.path = ln.frozen.Path
	} else if ln.parent != nil && ln.node != nil {
		ln.path = ln.parent.Path().Join(ln.node.SelfIdentity())
	} else if ln.node != nil && ln.node.Parent() == nil {
		ln.path = ln.node.SelfIdentity()
	} else if ln.dentry != nil {
		ln.path = ln.dentry.Path()
	}
}

func (ln *LiveNode) updateDentry(dentry *graphfs.CacheEntry) {
	if ln.dentry == dentry {
		return
	}

	defer ln.update()

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
}

func (ln *LiveNode) updateInode(ino *graphfs.INode) {
	if ln.inode == ino {
		return
	}

	defer ln.update()

	if ln.inode != nil {
		ln.inode.Unref()
		ln.inode = nil
	}

	if ino != nil {
		ln.inode = ino.Get()
	}
}

func (ln *LiveNode) updateOpenHandle(nh graphfs.NodeHandle) {
	defer ln.update()

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
	defer ln.update()

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

		if ln.frozen == nil {
			ln.g.markDirty(ln.node)
		}

		ln.flags |= liveNodeFlagHasInstance

		if ln.node.PsiNode() != nil {
			ln.flags |= liveNodeFlagInitialized
		} else {
			ln.flags &= ^liveNodeFlagInitialized
		}
	} else {
		ln.parent = nil

		ln.flags &= ^liveNodeFlagHasInstance
	}
}

func (ln *LiveNode) update() {
	ln.updatePath()

	ln.g.updateNodeCache(ln)
}

func (ln *LiveNode) Resolve(ctx context.Context, path psi.Path) (psi.Node, error) {
	if path.IsRelative() {
		path = ln.Path().Join(path)
	}

	return ln.g.ResolveNode(ctx, path)
}

func (ln *LiveNode) OnBeforeInitialize(node psi.Node) {
	ln.updateNode(node)
}

func (ln *LiveNode) OnAfterInitialize(node psi.Node) {
	ln.updateNode(node)
}

func (ln *LiveNode) OnAttributeChanged(key string, added any) {
	ln.flags |= liveNodeFlagDirty

	if ln.node != nil {
		ln.g.markDirty(ln.node)
	}
}

func (ln *LiveNode) OnAttributeRemoved(key string, removed any) {
	ln.flags |= liveNodeFlagDirty

	if ln.node != nil {
		ln.g.markDirty(ln.node)
	}
}

func (ln *LiveNode) OnEdgeAdded(added psi.Edge) {
	if le, ok := added.(*LiveEdge); ok && le.from == ln {
		ln.edges[le.key] = le
	} else {
		le := NewLiveEdge(ln, added.Key().GetKey())
		edge := le.ReplaceTo(added.To())

		ln.node.UpsertEdge(edge)
		ln.edges[le.key] = le
	}

	ln.markEdgeDirty(added.Key().GetKey())
}

func (ln *LiveNode) OnEdgeRemoved(removed psi.Edge) {
	ln.markEdgeDirty(removed.Key().GetKey())
}

func (ln *LiveNode) OnParentChange(newParent psi.Node) {
	if newParent == nil {
		ln.parent = nil
	} else {
		ln.parent = ln.g.nodeForNode(newParent)
	}

	ln.path = ln.parent.Path().Join(ln.node.SelfIdentity())

	if ln.frozen != nil && !ln.frozen.Path.Equals(ln.path) {
		ln.frozen = nil
	}

	ln.update()
}

func (ln *LiveNode) OnInvalidated() {
	ln.flags |= liveNodeFlagDirty

	if ln.node != nil {
		ln.g.markDirty(ln.node)
	}
}

func (ln *LiveNode) OnUpdated(ctx context.Context) error {
	return ln.Save(ctx)
}

func (ln *LiveNode) markEdgeDirty(key psi.EdgeKey) {
	e := ln.edges[key]

	if e == nil {
		e = NewLiveEdge(ln, key)

		ln.edges[key] = e
	}

	e.dirty = true

	ln.flags |= liveNodeFlagDirty

	if ln.node != nil {
		ln.g.markDirty(ln.node)
	}
}

func (ln *LiveNode) Lookup(key psi.EdgeReference) psi.Edge {
	k := key.GetKey()

	if le := ln.edges[k]; le != nil {
		return le
	}

	return nil
}

func (ln *LiveNode) Close() error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.node != nil {
		if closer, ok := ln.node.(io.Closer); ok {
			if err := closer.Close(); err != nil {
				return err
			}
		}

		ln.node = nil
	}

	if ln.handle != nil {
		if err := ln.handle.Close(); err != nil {
			return err
		}

		ln.handle = nil
	}

	if ln.inode != nil {
		ln.inode.Unref()
		ln.inode = nil
	}

	if ln.dentry != nil {
		ln.dentry.Unref()
		ln.dentry = nil
	}

	return nil
}
