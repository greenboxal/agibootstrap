package graphstore

import (
	"bytes"
	"context"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/multiformats/go-multihash"
)

type ObjectStore struct {
	ds datastore.Datastore
}

func NewObjectStore(ds datastore.Datastore) *ObjectStore {
	return &ObjectStore{ds: ds}
}

func (s *ObjectStore) prepareKey(contentId cid.Cid) datastore.Key {
	return datastore.KeyWithNamespaces([]string{"objects", contentId.String()})
}

func (s *ObjectStore) Get(ctx context.Context, contentId cid.Cid) (io.ReadCloser, error) {
	data, err := s.ds.Get(ctx, s.prepareKey(contentId))

	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *ObjectStore) Put(ctx context.Context, data io.Reader) (cid.Cid, error) {
	buf := new(bytes.Buffer)

	reader := io.TeeReader(data, buf)

	mh, err := multihash.SumStream(reader, multihash.SHA2_256, -1)

	if err != nil {
		return cid.Undef, err
	}

	contentId := cid.NewCidV1(cid.Raw, mh)

	if err := s.ds.Put(ctx, s.prepareKey(contentId), buf.Bytes()); err != nil {
		return cid.Undef, err
	}

	return contentId, nil
}

func (s *ObjectStore) Has(ctx context.Context, contentId cid.Cid) (bool, error) {
	return s.ds.Has(ctx, s.prepareKey(contentId))
}
