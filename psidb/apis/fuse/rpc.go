package fuse

import (
	"context"
	"io"
	"reflect"
	"syscall"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type psiRpcDir struct {
	fs.Inode

	core coreapi.Core
	path psi.Path

	getNode func(ctx context.Context) (any, error)
}

func NewPsiRpcDir(g coreapi.Core, path psi.Path, getNode func(ctx context.Context) (any, error)) fs.InodeEmbedder {
	return &psiRpcDir{
		core:    g,
		path:    path,
		getNode: getNode,
	}
}

func (pn *psiRpcDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if len(name) == 0 {
		return nil, syscall.ENOENT
	}

	definition := viewDefinitionsMap[name]

	if definition != nil {
		return pn.NewInode(ctx, NewPsiNodeFileView(pn.core, pn.path, definition.Prepare), fs.StableAttr{
			Mode: fuse.S_IFREG,
		}), 0
	}

	n, err := pn.getNode(ctx)

	if err != nil {
		if errors.Is(err, psi.ErrNodeNotFound) {
			return nil, syscall.ENOENT
		}

		logger.Error("failed to get node", "path", pn.path, "err", err)

		return nil, syscall.EIO
	}

	definitions := pn.buildDefinitions(n)

	for definitions.Next() {
		def := definitions.Value()

		if def.Name == name {
			return pn.NewInode(ctx, NewPsiNodeFileView(pn.core, pn.path, def.Prepare), fs.StableAttr{
				Mode: fuse.S_IFREG,
			}), 0
		}
	}

	return nil, syscall.ENOENT
}

func (pn *psiRpcDir) buildDefinitions(node any) iterators.Iterator[NodeViewDefinition] {
	if node == nil {
		return iterators.Empty[NodeViewDefinition]()
	}

	index := 0
	typ := reflect.TypeOf(node)

	if typ.Kind() != reflect.Ptr {
		typ = reflect.PtrTo(typ)
	}

	definitions := iterators.NewIterator(func() (def NodeViewDefinition, ok bool) {
		for {
			var inBuilder func(ctx context.Context, receiver any) []reflect.Value
			var outBuilder func([]reflect.Value) (any, error)

			if index >= typ.NumMethod() {
				ok = false
				return
			}

			m := typ.Method(index)
			index++

			switch m.Type.NumIn() {
			case 1:
				inBuilder = func(ctx context.Context, receiver any) []reflect.Value {
					return []reflect.Value{}
				}

			case 2:
				if m.Type.In(1) != contextType {
					continue
				}

				inBuilder = func(ctx context.Context, receiver any) []reflect.Value {
					return []reflect.Value{reflect.ValueOf(ctx)}
				}
			default:
				continue
			}

			switch m.Type.NumOut() {
			case 0:
				continue

			case 1:
				outBuilder = func(values []reflect.Value) (any, error) {
					return values[0].Interface(), nil
				}

				break

			case 2:
				if m.Type.Out(1) != errorType {
					continue
				}

				outBuilder = func(values []reflect.Value) (any, error) {
					e := values[1]

					if e.IsValid() && !e.IsNil() {
						return nil, e.Interface().(error)
					}

					return values[0].Interface(), nil
				}
			}

			def = NodeViewDefinition{}
			def.Name = m.Name
			def.Prepare = func(ctx context.Context, w io.Writer, n psi.Node) error {
				nv := reflect.ValueOf(n)
				mi := nv.MethodByName(m.Name)

				in := inBuilder(ctx, n)
				out := mi.Call(in)
				result, err := outBuilder(out)

				if err != nil {
					return err
				}

				if result == nil {
					return nil
				}

				return ipld.EncodeStreaming(w, typesystem.Wrap(result), dagjson.Encode)
			}

			ok = true

			return
		}
	})

	return definitions
}

func (pn *psiRpcDir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	node, err := pn.getNode(ctx)

	if err != nil {
		if errors.Is(err, psi.ErrNodeNotFound) {
			return nil, syscall.ENOENT
		}

		logger.Error("failed to get node", "path", pn.path, "err", err)

		return nil, syscall.EIO
	}

	definitions := pn.buildDefinitions(node)

	it := iterators.Map(definitions, func(def NodeViewDefinition) fuse.DirEntry {
		return fuse.DirEntry{
			Name: def.Name,
			Mode: fuse.S_IFREG,
		}
	})

	return fs.NewListDirStream(iterators.ToSlice(it)), 0
}

var _ fs.NodeLookuper = (*psiRpcDir)(nil)
var _ fs.NodeReaddirer = (*psiRpcDir)(nil)

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
var errorType = reflect.TypeOf((*context.Context)(nil)).Elem()
