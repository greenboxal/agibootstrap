package coreapi

import (
	"context"
	"encoding/hex"
	"os"
	"sync"

	badger2 "github.com/dgraph-io/badger"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/storage/dsadapter"

	"github.com/greenboxal/agibootstrap/psidb/core/refcount"
)

type DataStoreConfig interface {
	CreateDataStore(ctx context.Context) (DataStore, error)
}

type ExistingDataStore struct {
	DataStore
}

func (e ExistingDataStore) CreateDataStore(ctx context.Context) (DataStore, error) {
	return e, nil
}

func (e ExistingDataStore) Close() error {
	return nil
}

type sharedResourceCache[K comparable, V any] struct {
	mu        sync.RWMutex
	cache     map[K]*refcount.RefCounted[V]
	factory   func(k K) (V, error)
	finalizer func(v V) error
}

func newSharedResourceCache[K comparable, V any](factory func(k K) (V, error), finalizer func(v V) error) *sharedResourceCache[K, V] {
	return &sharedResourceCache[K, V]{
		cache:     make(map[K]*refcount.RefCounted[V]),
		factory:   factory,
		finalizer: finalizer,
	}
}

func (c *sharedResourceCache[K, V]) Get(key K) (refcount.ObjectHandle[V], error) {
	c.mu.RLock()
	if v, ok := c.cache[key]; ok {
		c.mu.RUnlock()
		return v.Ref(), nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.cache[key]; ok {
		return v.Ref(), nil
	}

	obj, err := c.factory(key)

	if err != nil {
		return nil, err
	}

	rc := &refcount.RefCounted[V]{}

	refcount.MakeRefCounted(rc, obj, func(obj refcount.ObjectSlot[V], value V, next refcount.ReferenceFinalizer[V]) {
		c.mu.Lock()
		defer c.mu.Unlock()

		if obj.IsValid() {
			return
		}

		if c.finalizer != nil {
			if err := c.finalizer(value); err != nil {
				panic(err)
			}
		}

		delete(c.cache, key)
	})

	c.cache[key] = rc

	return rc.Ref(), nil
}

var badgerSrc = newSharedResourceCache(func(k BadgerDataStoreConfig) (DataStore, error) {
	dsOpts := badger.DefaultOptions

	if err := os.MkdirAll(k.Path, 0755); err != nil {
		return nil, err
	}

	ds, err := badger.NewDatastore(k.Path, &dsOpts)

	if err != nil {
		return nil, err
	}

	dsa := &dsadapter.Adapter{
		Wrapped: ds,

		EscapingFunc: func(s string) string {
			return "_cas/" + hex.EncodeToString([]byte(s))
		},
	}

	lsys := cidlink.DefaultLinkSystem()
	lsys.SetReadStorage(dsa)
	lsys.SetWriteStorage(dsa)
	lsys.TrustedStorage = true

	return &BadgerDataStore{Batching: ds, lsys: lsys}, nil
}, func(v DataStore) error {
	return v.Close()
})

type BadgerDataStoreConfig struct {
	Path string `json:"path"`
}

func (b BadgerDataStoreConfig) CreateDataStore(ctx context.Context) (DataStore, error) {
	handle, err := badgerSrc.Get(b)

	if err != nil {
		return nil, err
	}

	return &refCountedMetadataStore{DataStore: handle.Get()}, nil
}

type refCountedMetadataStore struct {
	DataStore

	handle refcount.ObjectHandle[DataStore]
}

func (r *refCountedMetadataStore) Close() error {
	if r.handle == nil {
		return nil
	}

	r.handle.Release()
	r.handle = nil
	r.DataStore = nil

	return nil
}

type BadgerDataStore struct {
	datastore.Batching

	lsys linking.LinkSystem
}

func (m *BadgerDataStore) LinkSystem() *linking.LinkSystem {
	return &m.lsys
}

func (m *BadgerDataStore) DB() *badger2.DB {
	return m.Batching.(*badger.Datastore).DB
}
