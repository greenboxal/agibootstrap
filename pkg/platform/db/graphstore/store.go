package graphstore

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Store struct {
	os *ObjectStore

	ds datastore.Batching
}

func NewStore(ds datastore.Batching, os *ObjectStore) *Store {
	return &Store{ds: ds, os: os}
}

func (s *Store) UpsertNode(ctx context.Context, n psi.Node) (*FrozenNode, error) {
	batch, err := s.ds.Batch(ctx)

	if err != nil {
		return nil, err
	}

	fn, _, err := s.batchUpsertNode(ctx, batch, n)

	if err != nil {
		return nil, err
	}

	if err := batch.Commit(ctx); err != nil {
		return nil, err
	}

	return fn, nil
}

func (s *Store) batchUpsertNode(ctx context.Context, batch datastore.Batch, n psi.Node) (*FrozenNode, []*FrozenEdge, error) {
	data, id, err := SerializeNode(n)

	if err != nil {
		return nil, nil, err
	}

	if _, err := s.os.Put(ctx, bytes.NewReader(data)); err != nil {
		return nil, nil, err
	}

	fn := &FrozenNode{
		Cid:        cidlink.Link{Cid: id},
		UUID:       n.UUID(),
		Type:       n.PsiNodeType(),
		Version:    n.PsiNodeVersion(),
		Attributes: n.Attributes(),
	}

	data, err = ipld.Encode(typesystem.Wrap(fn), dagjson.Encode)

	if err != nil {
		return nil, nil, err
	}

	id, err = s.os.Put(ctx, bytes.NewReader(data))

	if err != nil {
		return nil, nil, err
	}

	headKey := fmt.Sprintf("refs/nodes/%s/HEAD", n.UUID())
	versionKey := fmt.Sprintf("refs/nodes/%s/%d", n.UUID(), n.PsiNodeVersion())

	if err := batch.Put(ctx, datastore.NewKey(headKey), id.Bytes()); err != nil {
		return nil, nil, err
	}

	if err := batch.Put(ctx, datastore.NewKey(versionKey), id.Bytes()); err != nil {
		return nil, nil, err
	}

	edges := make([]*FrozenEdge, 0)

	childIndex := int64(0)
	for it := n.ChildrenIterator(); it.Next(); childIndex++ {
		key := psi.EdgeKey{
			Kind:  psi.EdgeKindChild,
			Index: childIndex,
		}

		edge := psi.NewEdgeBase(key, n, it.Value())

		fe, feCid, err := s.batchUpsertEdge(ctx, batch, edge)

		if err != nil {
			return nil, nil, err
		}

		edgeKey := fmt.Sprintf("nodes/%s/%s", fn.Cid.String(), edge.Key().GetKey())

		if err := batch.Put(ctx, datastore.NewKey(edgeKey), feCid.Bytes()); err != nil {
			return nil, nil, err
		}

		edges = append(edges, fe)
	}

	for it := n.Edges(); it.Next(); {
		fe, _, err := s.batchUpsertEdge(ctx, batch, it.Edge())

		if err != nil {
			return nil, nil, err
		}

		edges = append(edges, fe)
	}

	return fn, edges, nil
}

func (s *Store) UpsertEdge(ctx context.Context, edge psi.Edge) (*FrozenEdge, error) {
	batch, err := s.ds.Batch(ctx)

	if err != nil {
		return nil, err
	}

	fe, _, err := s.batchUpsertEdge(ctx, batch, edge)

	if err != nil {
		return nil, err
	}

	if err := batch.Commit(ctx); err != nil {
		return nil, err
	}

	return fe, nil
}

func (s *Store) batchUpsertEdge(ctx context.Context, batch datastore.Batch, edge psi.Edge) (*FrozenEdge, cid.Cid, error) {
	data, err := ipld.Encode(typesystem.Wrap(edge), dagjson.Encode)

	if err != nil {
		return nil, cid.Undef, err
	}

	contentId, err := s.os.Put(ctx, bytes.NewReader(data))

	if err != nil {
		return nil, cid.Undef, err
	}

	fe := &FrozenEdge{
		Cid:  cidlink.Link{Cid: contentId},
		Key:  edge.Key().GetKey(),
		From: edge.From().UUID(),
		To:   edge.To().UUID(),
	}

	data, err = ipld.Encode(typesystem.Wrap(fe), dagjson.Encode)

	if err != nil {
		return nil, cid.Undef, err
	}

	contentId, err = s.os.Put(ctx, bytes.NewReader(data))

	if err != nil {
		return nil, cid.Undef, err
	}

	key := fmt.Sprintf("edges/%s/%s", edge.From().UUID(), edge.Key().GetKey())

	if err := batch.Put(ctx, datastore.NewKey(key), contentId.Bytes()); err != nil {
		return nil, cid.Undef, err
	}

	return fe, contentId, nil
}

func (s *Store) LoadNode(ctx context.Context, fn *FrozenNode) (psi.Node, error) {
	reader, err := s.os.Get(ctx, fn.Cid.Cid)

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	n, err := DeserializeNode(fn.UUID, data)

	if err != nil {
		return nil, err
	}

	for k, v := range fn.Attributes {
		n.SetAttribute(k, v)
	}

	return n, err
}

func (s *Store) GetEdgeByCid(ctx context.Context, id cid.Cid) (*FrozenEdge, error) {
	reader, err := s.os.Get(ctx, id)

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	n, err := ipld.DecodeUsingPrototype(data, dagjson.Decode, frozenEdgeType.IpldPrototype())

	if err != nil {
		return nil, err
	}

	return typesystem.Unwrap(n).(*FrozenEdge), nil
}

func (s *Store) GetNodeByCid(ctx context.Context, id cid.Cid) (*FrozenNode, error) {
	reader, err := s.os.Get(ctx, id)

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(reader)

	if err != nil {
		return nil, err
	}

	n, err := ipld.DecodeUsingPrototype(data, dagjson.Decode, frozenNodeType.IpldPrototype())

	if err != nil {
		return nil, err
	}

	return typesystem.Unwrap(n).(*FrozenNode), nil
}

func (s *Store) GetNodeByID(ctx context.Context, id psi.NodeID, version int64) (*FrozenNode, error) {
	var key string

	if version == -1 {
		key = fmt.Sprintf("refs/nodes/%s/HEAD", id)
	} else {
		key = fmt.Sprintf("refs/nodes/%s/%d", id, version)
	}

	cidBytes, err := s.ds.Get(ctx, datastore.NewKey(key))

	if err != nil {
		return nil, err
	}

	contentId, err := cid.Cast(cidBytes)

	if err != nil {
		return nil, err
	}

	return s.GetNodeByCid(ctx, contentId)
}

func (s *Store) GetNodeEdges(ctx context.Context, id psi.NodeID, version int64) (iterators.Iterator[*FrozenEdge], error) {
	var q query.Query

	if version == -1 {
		q.Prefix = fmt.Sprintf("edges/%s/", id)
	} else {
		n, err := s.GetNodeByID(ctx, id, version)

		if err != nil {
			return nil, err
		}

		q.Prefix = fmt.Sprintf("nodes/%s/%d", n.Cid.String(), version)
	}

	it, err := s.ds.Query(ctx, q)

	if err != nil {
		return nil, err
	}

	return iterators.NewIterator(func() (*FrozenEdge, bool) {
		res, ok := it.NextSync()

		if !ok {
			return nil, false
		}

		contentId, err := cid.Cast(res.Value)

		if err != nil {
			return nil, false
		}

		fe, err := s.GetEdgeByCid(ctx, contentId)

		if err != nil {
			return nil, false
		}

		return fe, true
	}), nil

}

func (s *Store) ResolvePath(ctx context.Context, path psi.Path) (psi.NodeID, error) {
	return "", nil
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
