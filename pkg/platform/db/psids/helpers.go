package psids

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type TypedKey[T any] interface {
	Key() datastore.Key
	String() string

	Serialize(value T) ([]byte, error)
	Deserialize(data []byte) (T, error)
}

type typedKey[T any] datastore.Key

func (tk typedKey[T]) Key() datastore.Key { return datastore.Key(tk) }

func (tk typedKey[T]) String() string { return datastore.Key(tk).String() }

func (tk typedKey[T]) Serialize(value T) ([]byte, error) {
	return ipld.Encode(typesystem.Wrap(value), dagcbor.Encode)
}

func (tk typedKey[T]) Deserialize(data []byte) (result T, err error) {
	n, err := ipld.DecodeUsingPrototype(data, dagcbor.Decode, typesystem.TypeOf((*T)(nil)).IpldPrototype())

	if err != nil {
		return result, err
	}

	result, ok := typesystem.TryUnwrap[T](n)

	if !ok {
		return result, errors.New("unexpected node type")
	}

	return result, nil
}

type keyWithEncoding[T any] struct {
	typedKey[T]

	encoder ipld.Encoder
	decoder ipld.Decoder
}

func (tk keyWithEncoding[T]) Serialize(value T) ([]byte, error) {
	return ipld.Encode(typesystem.Wrap(value), tk.encoder)
}

func (tk keyWithEncoding[T]) Deserialize(data []byte) (result T, err error) {
	n, err := ipld.Decode(data, tk.decoder)

	if err != nil {
		return result, err
	}

	result, ok := typesystem.TryUnwrap[T](n)

	if !ok {
		return result, errors.New("unexpected node type")
	}

	return result, nil
}

type uintKey struct{ typedKey[uint64] }

func (tk uintKey) Serialize(value uint64) ([]byte, error) {
	return []byte(strconv.FormatUint(value, 10)), nil
}

func (tk uintKey) Deserialize(data []byte) (uint64, error) {
	return strconv.ParseUint(string(data), 10, 64)
}

type stringKey struct{ typedKey[string] }

func (tk stringKey) Serialize(value string) ([]byte, error) { return ([]byte)(value), nil }

func (tk stringKey) Deserialize(data []byte) (string, error) { return string(data), nil }

type bytesKey struct{ typedKey[[]byte] }

func (tk bytesKey) Serialize(value []byte) ([]byte, error) { return value, nil }

func (tk bytesKey) Deserialize(data []byte) ([]byte, error) { return data, nil }

func Key[T any](segments ...string) TypedKey[T] {
	typ := reflect.TypeOf((*T)(nil)).Elem()

	if typ.Kind() == reflect.String {
		k := stringKey{typedKey[string](datastore.KeyWithNamespaces(segments))}

		return any(k).(TypedKey[T])
	} else if typ.Kind() == reflect.Slice && typ.Elem().Kind() == reflect.Uint8 {
		k := bytesKey{typedKey[[]byte](datastore.KeyWithNamespaces(segments))}

		return any(k).(TypedKey[T])
	} else if typ.Kind() == reflect.Uint64 {
		k := uintKey{typedKey[uint64](datastore.KeyWithNamespaces(segments))}

		return any(k).(TypedKey[T])
	}

	return typedKey[T](datastore.KeyWithNamespaces(segments))
}

func KeyTemplate[T any](format string) func(args ...any) TypedKey[T] {
	return func(args ...any) TypedKey[T] {
		return Key[T](fmt.Sprintf(format, args...))
	}
}

func WithEncoding[T any](key TypedKey[T], encoder ipld.Encoder, decoder ipld.Decoder) TypedKey[T] {
	return keyWithEncoding[T]{key.(typedKey[T]), encoder, decoder}
}

func Get[T any](ctx context.Context, ds datastore.Read, key TypedKey[T]) (result T, err error) {
	data, err := ds.Get(ctx, key.Key())

	if err != nil {
		if err == datastore.ErrNotFound {
			return result, psi.ErrNodeNotFound
		}

		return result, err
	}

	return key.Deserialize(data)
}

func GetFirst[T any](ctx context.Context, ds datastore.Read, prefix TypedKey[T], eval func(T) bool) (empty T, err error) {
	it, err := List(ctx, ds, prefix)

	if err != nil {
		if err == datastore.ErrNotFound {
			return empty, psi.ErrNodeNotFound
		}

		return empty, err
	}

	for it.Next() {
		v := it.Value()

		if eval(v) {
			return v, nil
		}
	}

	return empty, psi.ErrNodeNotFound
}

func GetOrDefault[T any](ctx context.Context, ds datastore.Read, key TypedKey[T], defaultValue T) (T, error) {
	result, err := Get(ctx, ds, key)

	if err != nil {
		if err == psi.ErrNodeNotFound {
			return defaultValue, nil
		}

		return defaultValue, err
	}

	return result, nil
}

func List[T any](ctx context.Context, ds datastore.Read, key TypedKey[T]) (iterators.Iterator[T], error) {
	return Query(ctx, ds, key, query.Query{})
}

func Query[T any](ctx context.Context, ds datastore.Read, key TypedKey[T], q query.Query) (iterators.Iterator[T], error) {
	q.Prefix = key.String()

	it, err := ds.Query(ctx, q)

	if err != nil {
		return nil, err
	}

	return iterators.NewIterator(func() (def T, ok bool) {
		res, ok := it.NextSync()

		if !ok {
			return def, false
		}

		v, err := key.Deserialize(res.Value)

		if err != nil {
			return def, false
		}

		return v, true
	}), nil
}

func Put[T any](ctx context.Context, ds datastore.Write, key TypedKey[T], value T) error {
	data, err := key.Serialize(value)

	if err != nil {
		return err
	}

	return ds.Put(ctx, key.Key(), data)
}

func Delete[T any](ctx context.Context, ds datastore.Write, key TypedKey[T]) error {
	return ds.Delete(ctx, key.Key())
}

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}
