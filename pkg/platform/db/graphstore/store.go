package graphstore

import (
	"context"
	"encoding/hex"
	"fmt"
	"path"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/dsadapter"

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
			return hex.EncodeToString([]byte(s))
		},
	}

	s.lsys = cidlink.DefaultLinkSystem()
	s.lsys.SetReadStorage(adapter)
	s.lsys.SetWriteStorage(adapter)
	s.lsys.TrustedStorage = true

	return s
}

func (s *Store) FreezeNode(ctx context.Context, n psi.Node) (*psi.FrozenNode, []*psi.FrozenEdge, ipld.Link, error) {
	fn := &psi.FrozenNode{
		Path:       n.CanonicalPath(),
		Type:       n.PsiNodeType().Name(),
		Version:    n.PsiNodeVersion(),
		Attributes: n.Attributes(),
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

		edge := psi.NewEdgeBase(key, n, it.Value())

		fe, feLink, err := s.FreezeEdge(ctx, edge)

		if err != nil {
			return nil, nil, nil, err
		}

		fn.Edges = append(fn.Edges, feLink.(cidlink.Link))

		edges = append(edges, fe)
	}

	for it := n.Edges(); it.Next(); {
		fe, feLink, err := s.FreezeEdge(ctx, it.Edge())

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
	if le, ok := edge.(*lazyEdge); ok && le.frozen != nil {
		if le.link == nil {
			link, err := s.lsys.Store(ipld.LinkContext{Ctx: ctx}, defaultLinkPrototype, typesystem.Wrap(le.frozen))

			if err != nil {
				return nil, nil, err
			}

			le.link = link
		}

		return le.frozen, le.link, nil
	}

	dataLink, err := s.lsys.Store(ipld.LinkContext{Ctx: ctx}, defaultLinkPrototype, typesystem.Wrap(edgeWrapper{Edge: edge}))

	if err != nil {
		return nil, nil, err
	}

	fe := &psi.FrozenEdge{
		Data:     dataLink.(cidlink.Link),
		Key:      edge.Key().GetKey(),
		FromPath: edge.From().CanonicalPath(),
		ToIndex:  edge.To().ID(),
	}

	for parent := edge.To(); parent != nil; parent = parent.Parent() {
		if parent == s.root {
			p := edge.To().CanonicalPath()
			fe.ToPath = &p
			break
		}
	}

	if toSnapshot := psi.GetNodeSnapshot(edge.To()); toSnapshot != nil {
		if l, ok := toSnapshot.CommittedLink().(cidlink.Link); ok {
			fe.ToLink = &l
		}
	}

	link, err := s.lsys.Store(ipld.LinkContext{Ctx: ctx}, defaultLinkPrototype, typesystem.Wrap(fe))

	if err != nil {
		return nil, nil, err
	}

	return fe, link, nil
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

func (s *Store) IndexNode(ctx context.Context, batch datastore.Batch, root string, fn *psi.FrozenNode, link ipld.Link) error {
	headKey := fmt.Sprintf("refs/heads/%s/%s", root, fn.Path)

	if err := batch.Put(ctx, datastore.NewKey(headKey), []byte(link.Binary())); err != nil {
		return err
	}

	return nil
}

func (s *Store) IndexEdge(ctx context.Context, batch datastore.Batch, root string, fe *psi.FrozenEdge, feLink cidlink.Link) error {
	key := fmt.Sprintf("refs/edges/%s/%s!/%s", root, fe.FromPath, fe.Key.AsPathElement())

	if err := batch.Put(ctx, datastore.NewKey(key), []byte(feLink.Binary())); err != nil {
		return err
	}

	return nil
}

func (s *Store) GetNodeByPath(ctx context.Context, root string, path psi.Path) (*psi.FrozenNode, ipld.Link, error) {
	key := fmt.Sprintf("refs/heads/%s/%s", root, path)

	cidBytes, err := s.ds.Get(ctx, datastore.NewKey(key))

	if err != nil {
		return nil, nil, err
	}

	contentId, err := cid.Cast(cidBytes)

	if err != nil {
		return nil, nil, err
	}

	link := cidlink.Link{Cid: contentId}
	fn, err := s.GetNodeByCid(ctx, link)

	if err != nil {
		return nil, nil, err
	}

	return fn, link, nil
}

func (s *Store) ListNodeEdges(ctx context.Context, root string, nodePath psi.Path) (iterators.Iterator[*psi.FrozenEdge], error) {
	var q query.Query

	q.Prefix = path.Join("/refs/edges", root, nodePath.String()+"!")

	it, err := s.ds.Query(ctx, q)

	if err != nil {
		return nil, err
	}

	return iterators.NewIterator(func() (*psi.FrozenEdge, bool) {
		res, ok := it.NextSync()

		if !ok {
			return nil, false
		}

		contentId, err := cid.Cast(res.Value)

		if err != nil {
			return nil, false
		}

		fe, err := s.GetEdgeByCid(ctx, cidlink.Link{Cid: contentId})

		if err != nil {
			return nil, false
		}

		return fe, true
	}), nil

}

func (s *Store) RemoveNode(ctx context.Context, path psi.Path) error {
	batch, err := s.ds.Batch(ctx)

	if err != nil {
		return err
	}

	if err := s.batchRemoveNode(ctx, batch, path); err != nil {
		return err
	}

	if err := batch.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Store) batchRemoveNode(ctx context.Context, batch datastore.Batch, path psi.Path) error {
	return nil
}

func (s *Store) RemoveEdge(ctx context.Context, nodeId psi.NodeID, key psi.EdgeKey) error {
	batch, err := s.ds.Batch(ctx)

	if err != nil {
		return err
	}

	if err := s.batchRemoveEdge(ctx, batch, nodeId, key); err != nil {
		return err
	}

	if err := batch.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Store) batchRemoveEdge(ctx context.Context, batch datastore.Batch, nodeId psi.NodeID, key psi.EdgeKey) error {
	return nil
}
