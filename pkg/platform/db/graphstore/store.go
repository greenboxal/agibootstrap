package graphstore

import (
	"context"
	"encoding/hex"

	"github.com/ipfs/go-datastore"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/dsadapter"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/psids"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Store struct {
	ds  datastore.Batching
	wal *WriteAheadLog

	lsys ipld.LinkSystem
	root psi.Node
}

func NewStore(ds datastore.Batching, wal *WriteAheadLog, root psi.Node) *Store {
	s := &Store{ds: ds, wal: wal, root: root}

	adapter := &dsadapter.Adapter{
		Wrapped: ds,

		EscapingFunc: func(s string) string {
			return "_cas/" + hex.EncodeToString([]byte(s))
		},
	}

	s.lsys = cidlink.DefaultLinkSystem()
	s.lsys.SetReadStorage(adapter)
	s.lsys.SetWriteStorage(adapter)
	s.lsys.TrustedStorage = true

	return s
}

func (s *Store) GetEdgeByCid(ctx context.Context, link ipld.Link) (*psi.FrozenEdge, error) {
	n, err := s.lsys.Load(ipld.LinkContext{Ctx: ctx}, link, frozenEdgeType.IpldPrototype())

	if err != nil {
		return nil, err
	}

	fe := typesystem.Unwrap(n).(psi.FrozenEdge)

	return &fe, nil
}

func (s *Store) GetNodeByCid(ctx context.Context, link ipld.Link) (*psi.FrozenNode, error) {
	n, err := s.lsys.Load(ipld.LinkContext{Ctx: ctx}, link, frozenNodeType.IpldPrototype())

	if err != nil {
		return nil, err
	}

	fn := typesystem.Unwrap(n).(psi.FrozenNode)

	return &fn, nil
}
func (s *Store) GetNodeByPath(ctx context.Context, path psi.Path) (*psi.FrozenNode, ipld.Link, error) {
	link, err := psids.Get(ctx, s.ds, dsKeyNodeHead(path))

	if err != nil {
		return nil, nil, err
	}

	fn, err := s.GetNodeByCid(ctx, link)

	if err != nil {
		return nil, nil, err
	}

	return fn, link, nil
}

func (s *Store) ListNodeEdges(ctx context.Context, nodePath psi.Path) (iterators.Iterator[*psi.FrozenEdge], error) {
	it, err := psids.List(ctx, s.ds, dsKeyEdgePrefix(nodePath))

	if err != nil {
		return nil, err
	}

	return iterators.NewIterator(func() (*psi.FrozenEdge, bool) {
		if !it.Next() {
			return nil, false
		}

		fe, err := s.GetEdgeByCid(ctx, it.Value())

		if err != nil {
			return nil, false
		}

		return fe, true
	}), nil
}

func (s *Store) IndexNode(ctx context.Context, batch datastore.Batch, fn *psi.FrozenNode, link ipld.Link) error {
	if err := psids.Put(ctx, batch, dsKeyNodeHead(fn.Path), link.(cidlink.Link)); err != nil {
		return err
	}

	return nil
}

func (s *Store) IndexEdge(ctx context.Context, batch datastore.Batch, fe *psi.FrozenEdge, feLink cidlink.Link) error {
	key := dsKeyEdge(fe.FromPath, fe.Key.AsPathElement())

	return psids.Put(ctx, batch, key, feLink)
}

func (s *Store) RemoveNodeFromIndex(ctx context.Context, batch datastore.Batch, path psi.Path) error {
	return psids.Delete(ctx, batch, dsKeyNodeHead(path))
}

func (s *Store) RemoveEdgeFromIndex(ctx context.Context, batch datastore.Batch, path psi.Path, key psi.EdgeKey) error {
	return psids.Delete(ctx, batch, dsKeyEdge(path, key))
}

func (s *Store) FreezeNode(ctx context.Context, n psi.Node) (*psi.FrozenNode, []*psi.FrozenEdge, ipld.Link, error) {
	fn := &psi.FrozenNode{
		Path:    n.CanonicalPath(),
		Type:    n.PsiNodeType().Name(),
		Version: n.PsiNodeVersion(),
	}

	edges := make([]*psi.FrozenEdge, 0)

	childIndex := int64(0)

	if !n.PsiNodeType().Definition().IsRuntimeOnly {
		dataLink, err := s.lsys.Store(ipld.LinkContext{Ctx: ctx}, defaultLinkPrototype, typesystem.Wrap(n))

		if err != nil {
			return nil, nil, nil, err
		}

		l := dataLink.(cidlink.Link)
		fn.Data = &l
	}

	for it := n.ChildrenIterator(); it.Next(); childIndex++ {
		key := psi.EdgeKey{
			Kind:  psi.EdgeKindChild,
			Index: childIndex,
		}

		if named, ok := it.Value().(psi.NamedNode); ok {
			key.Name = named.PsiNodeName()
		}

		edge := psi.NewSimpleEdge(key, n, it.Value())

		fe, feLink, err := s.FreezeEdge(ctx, edge)

		if err != nil {
			return nil, nil, nil, err
		}

		fn.Edges = append(fn.Edges, feLink.(cidlink.Link))

		edges = append(edges, fe)
	}

	for it := n.Edges(); it.Next(); {
		fe, feLink, err := s.FreezeEdge(ctx, it.Value())

		if err != nil {
			return nil, nil, nil, err
		}

		fn.Edges = append(fn.Edges, feLink.(cidlink.Link))

		edges = append(edges, fe)
	}

	link, err := s.lsys.Store(ipld.LinkContext{Ctx: ctx}, defaultLinkPrototype, typesystem.Wrap(fn))

	if err != nil {
		return nil, nil, nil, err
	}

	return fn, edges, link, nil
}

func (s *Store) FreezeEdge(ctx context.Context, edge psi.Edge) (*psi.FrozenEdge, ipld.Link, error) {
	snap := edge.PsiEdgeBase().GetSnapshot()

	if snap.Frozen != nil {
		if snap.Link == nil {
			link, err := s.lsys.Store(ipld.LinkContext{Ctx: ctx}, defaultLinkPrototype, typesystem.Wrap(snap.Frozen))

			if err != nil {
				return nil, nil, err
			}

			snap.Link = link
		}

		return snap.Frozen, snap.Link, nil
	}

	dataLink, err := s.lsys.Store(ipld.LinkContext{Ctx: ctx}, defaultLinkPrototype, typesystem.Wrap(edgeWrapper{Edge: edge}))

	if err != nil {
		return nil, nil, err
	}

	if edge == nil || edge.From() == nil {
		panic("wtf")
	}

	dl := dataLink.(cidlink.Link)
	fe := &psi.FrozenEdge{
		Data:     &dl,
		Key:      edge.Key().GetKey(),
		FromPath: edge.From().CanonicalPath(),
		ToIndex:  edge.To().ID(),
	}

	to := edge.To()

	if to == nil {
		fe.ToIndex = -1
	} else {
		fe.ToIndex = to.ID()

		toPath := to.CanonicalPath()

		if toPath.IsChildOf(s.root.CanonicalPath()) {
			fe.ToPath = &toPath
		}
	}

	if toSnapshot := psi.GetNodeSnapshot(edge.To()); toSnapshot != nil {
		if l, ok := toSnapshot.CommitLink().(cidlink.Link); ok {
			fe.ToLink = &l
		}
	}

	link, err := s.lsys.Store(ipld.LinkContext{Ctx: ctx}, defaultLinkPrototype, typesystem.Wrap(fe))

	if err != nil {
		return nil, nil, err
	}

	edge.PsiEdgeBase().SetSnapshot(psi.EdgeSnapshot{
		Frozen: fe,
		Link:   link,
	})

	return fe, link, nil
}
