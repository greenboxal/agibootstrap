package psifuse

import (
	"bytes"
	"context"
	"io"
	"sync"
	"syscall"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

var logger = logging.GetLogger("psifuse")

func NewPsiNodeWrapper(g *graphstore.IndexedGraph, path psi.Path) fs.InodeEmbedder {
	return &psiNode{
		g:    g,
		path: path,
	}
}

type psiNode struct {
	fs.Inode

	g    *graphstore.IndexedGraph
	path psi.Path
}

var _ fs.NodeLookuper = (*psiNode)(nil)
var _ fs.NodeReadlinker = (*psiNode)(nil)
var _ fs.NodeReaddirer = (*psiNode)(nil)

func (pn *psiNode) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	return []byte(pn.path.String()), 0
}

func (pn *psiNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	r := make([]fuse.DirEntry, 0, 2)

	r = append(r, fuse.DirEntry{
		Name: "!node.json",
		Mode: fuse.S_IFREG,
	})
	r = append(r, fuse.DirEntry{
		Name: "!snapshot.json",
		Mode: fuse.S_IFREG,
	})

	edges, err := pn.g.Store().ListNodeEdges(ctx, pn.g.Root().UUID(), pn.path)

	if err != nil {
		if !errors.Is(err, psi.ErrNodeNotFound) {
			logger.Error("failed to list node edges", "path", pn.path, "err", err)
		}

		return nil, syscall.EIO
	}

	for edges.Next() {
		edge := edges.Value()
		k := edge.Key

		if k.Kind == psi.EdgeKindChild && k.Name != "" {
			k.Index = 0
		}

		d := fuse.DirEntry{
			Name: k.String(),
			Ino:  0x8800000000000000 | uint64(edge.ToIndex),
		}

		if k.Kind == psi.EdgeKindChild {
			d.Mode = fuse.S_IFDIR
		} else {
			d.Mode = fuse.S_IFLNK
		}

		r = append(r, d)
	}

	return fs.NewListDirStream(r), 0
}

func (pn *psiNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	switch name {
	case "!node.json":
		return pn.NewInode(ctx, NewPsiNodeFileView(pn.g, pn.path, func(w io.Writer, n psi.Node) error {
			return ipld.EncodeStreaming(w, typesystem.Wrap(n), dagjson.Encode)
		}), fs.StableAttr{
			Mode: fuse.S_IFREG,
		}), 0

	case "!snapshot.json":
		return pn.NewInode(ctx, NewPsiNodeFileView(pn.g, pn.path, func(w io.Writer, n psi.Node) error {
			snap := psi.GetNodeSnapshot(n)

			if snap == nil {
				return nil
			}

			return ipld.EncodeStreaming(w, typesystem.Wrap(snap.FrozenNode()), dagjson.Encode)
		}), fs.StableAttr{
			Mode: fuse.S_IFREG,
		}), 0

	case "!node.txt":
		return pn.NewInode(ctx, NewPsiNodeFileView(pn.g, pn.path, func(w io.Writer, n psi.Node) error {
			w.Write([]byte(n.String()))
			return nil
		}), fs.StableAttr{
			Mode: fuse.S_IFREG,
		}), 0

	default:
		return pn.lookupNode(ctx, name, out)
	}
}

func (pn *psiNode) lookupNode(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
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
		Ino:  0x8800000000000000 | uint64(n.ID()),
	}

	operations := NewPsiNodeWrapper(pn.g, path)

	child := pn.NewInode(ctx, operations, stable)

	return child, 0
}

type psiNodeFileView struct {
	fs.Inode

	g    *graphstore.IndexedGraph
	path psi.Path

	prepare func(w io.Writer, n psi.Node) error
}

func NewPsiNodeFileView(g *graphstore.IndexedGraph, path psi.Path, prepare func(w io.Writer, n psi.Node) error) *psiNodeFileView {
	return &psiNodeFileView{
		g:       g,
		path:    path,
		prepare: prepare,
	}
}

type psiNodeFileHandle struct {
	view *psiNodeFileView

	mu sync.Mutex

	node         psi.Node
	data         bytes.Buffer
	materialized bool
}

func (nfw *psiNodeFileView) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if fuseFlags&(syscall.O_RDWR|syscall.O_WRONLY) != 0 {
		return nil, 0, syscall.EROFS
	}

	n, err := nfw.g.ResolveNode(ctx, nfw.path)

	if err != nil {
		if !errors.Is(err, psi.ErrNodeNotFound) {
			logger.Errorw("failed to resolve node", "path", nfw.path, "err", err)
		}

		return nil, 0, syscall.ENOENT
	}

	fh = &psiNodeFileHandle{view: nfw, node: n}

	return fh, fuse.FOPEN_DIRECT_IO, 0
}

func (p *psiNodeFileHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	data, err := p.materialize(ctx)

	if err != nil {
		return nil, syscall.EIO
	}

	if off >= int64(data.Len()) {
		return fuse.ReadResultData(nil), 0
	}

	end := off + int64(len(dest))

	if end > int64(data.Len()) {
		end = int64(data.Len())
	}

	// We could copy to the `dest` buffer, but since we have a
	// []byte already, return that.
	return fuse.ReadResultData(data.Bytes()[off:end]), 0
}

func (p *psiNodeFileHandle) materialize(ctx context.Context) (*bytes.Buffer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.materialized {
		return &p.data, nil
	}

	p.data.Reset()

	if err := p.view.prepare(&p.data, p.node); err != nil {
		return nil, err
	}

	p.materialized = true

	return &p.data, nil
}

var _ fs.NodeOpener = (*psiNodeFileView)(nil)
var _ fs.FileReader = (*psiNodeFileHandle)(nil)
