package psifuse

import (
	"context"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type psiNodeDir struct {
	fs.Inode

	g    *graphstore.IndexedGraph
	path psi.Path
}

var logger = logging.GetLogger("psifuse")

func NewPsiNodeDir(g *graphstore.IndexedGraph, path psi.Path) fs.InodeEmbedder {
	return &psiNodeDir{
		g:    g,
		path: path,
	}
}

func (pn *psiNodeDir) getNode(ctx context.Context) (psi.Node, error) {
	return pn.g.ResolveNode(ctx, pn.path)
}

func (pn *psiNodeDir) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if len(name) == 0 {
		return nil, syscall.ENOENT
	}

	if name == "!RPC" {
		return pn.NewInode(ctx, NewPsiRpcDir(pn.g, pn.path, func(ctx context.Context) (any, error) {
			return pn.getNode(ctx)
		}), fs.StableAttr{
			Mode: fuse.S_IFDIR,
		}), 0
	}

	if name[0] == '!' {
		definition := viewDefinitionsMap[name]

		if definition != nil {
			return pn.NewInode(ctx, NewPsiNodeFileView(pn.g, pn.path, definition.Prepare), fs.StableAttr{
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
				return pn.NewInode(ctx, NewPsiNodeFileView(pn.g, pn.path, def.Prepare), fs.StableAttr{
					Mode: fuse.S_IFREG,
				}), 0
			}
		}

		return nil, syscall.ENOENT
	}

	return pn.lookupNode(ctx, name, out)
}

func (pn *psiNodeDir) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	n, err := pn.getNode(ctx)

	if err != nil {
		if errors.Is(err, psi.ErrNodeNotFound) {
			return nil, syscall.ENOENT
		}

		logger.Error("failed to get node", "path", pn.path, "err", err)

		return nil, syscall.EIO
	}

	return []byte(n.CanonicalPath().String()), 0
}

func (pn *psiNodeDir) buildDefinitions(node psi.Node) iterators.Iterator[NodeViewDefinition] {
	return iterators.FromSlice(viewDefinitions)
}

func (pn *psiNodeDir) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	node, _ := pn.getNode(ctx)
	definitions := pn.buildDefinitions(node)

	edges, err := pn.g.Store().ListNodeEdges(ctx, pn.path)

	if err != nil {
		if errors.Is(err, psi.ErrNodeNotFound) {
			return nil, syscall.ENOENT
		}

		logger.Error("failed to list node edges", "path", pn.path, "err", err)

		return nil, syscall.EIO
	}

	staticIterator := iterators.Map(definitions, func(def NodeViewDefinition) fuse.DirEntry {
		return fuse.DirEntry{
			Name: def.Name,
			Mode: fuse.S_IFREG,
		}
	})

	edgesIterator := iterators.Map(edges, func(edge *psi.FrozenEdge) fuse.DirEntry {
		k := edge.Key

		if k.Kind == psi.EdgeKindChild && k.Name != "" {
			k.Index = 0
		}

		d := fuse.DirEntry{
			Name: k.String(),
		}

		if k.Kind == psi.EdgeKindChild {
			d.Mode = fuse.S_IFDIR
			d.Ino = 0x8100000000000000 | uint64(edge.ToIndex)
		} else {
			d.Mode = fuse.S_IFLNK
		}

		return d
	})

	rpcIterator := iterators.Single(fuse.DirEntry{
		Name: "!RPC",
		Mode: fuse.S_IFDIR,
	})

	all := iterators.Concat(rpcIterator, staticIterator, edgesIterator)

	return fs.NewListDirStream(iterators.ToSlice(all)), 0
}

func (pn *psiNodeDir) lookupNode(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	p, err := psi.ParsePathElement(name)

	if err != nil {
		logger.Errorw("failed to parse path element", "name", name, "err", err)
		return nil, syscall.ENOENT
	}

	path := pn.path.Child(p)
	n, err := pn.g.ResolveNode(ctx, path)

	if err != nil {
		if !errors.Is(err, psi.ErrNodeNotFound) {
			logger.Errorw("failed to resolve node", "path", path, "err", err)
		}

		return nil, syscall.ENOENT
	}

	stable := fs.StableAttr{
		Mode: fuse.S_IFDIR,
		Ino:  0x8100000000000000 | uint64(n.ID()),
	}

	operations := NewPsiNodeDir(pn.g, path)

	child := pn.NewInode(ctx, operations, stable)

	return child, 0
}

var _ fs.NodeLookuper = (*psiNodeDir)(nil)
var _ fs.NodeReadlinker = (*psiNodeDir)(nil)
var _ fs.NodeReaddirer = (*psiNodeDir)(nil)