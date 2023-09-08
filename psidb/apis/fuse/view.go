package fuse

import (
	"bytes"
	"context"
	"io"
	"sync"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/psidb/typesystem"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type NodeViewFilterFunc func(ctx context.Context, n psi.Node) bool
type NodeViewPrepareFunc func(ctx context.Context, w io.Writer, n psi.Node) error

type NodeViewDefinition struct {
	Name    string
	Filter  NodeViewFilterFunc
	Prepare NodeViewPrepareFunc
}

var viewDefinitions = []NodeViewDefinition{
	{
		Name: "!node.json",
		Prepare: func(ctx context.Context, w io.Writer, n psi.Node) error {
			return ipld.EncodeStreaming(w, typesystem.Wrap(n), dagjson.Encode)
		},
	},
}

var viewDefinitionsMap = iterators.ToMap(iterators.Map(iterators.FromSlice(viewDefinitions), func(def NodeViewDefinition) iterators.KeyValue[string, *NodeViewDefinition] {
	return iterators.KeyValue[string, *NodeViewDefinition]{
		K: def.Name,
		V: &def,
	}
}))

type psiNodeFileView struct {
	fs.Inode

	core coreapi.Core
	path psi.Path

	prepare NodeViewPrepareFunc
}

func NewPsiNodeFileView(core coreapi.Core, path psi.Path, prepare NodeViewPrepareFunc) *psiNodeFileView {
	return &psiNodeFileView{
		core:    core,
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

	tx coreapi.Transaction
}

func (nfw *psiNodeFileView) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if fuseFlags&(syscall.O_RDWR|syscall.O_WRONLY) != 0 {
		return nil, 0, syscall.EROFS
	}

	tx, err := nfw.core.BeginTransaction(ctx)

	if err != nil {
		return nil, 0, syscall.EAGAIN
	}

	n, err := tx.Resolve(ctx, nfw.path)

	if err != nil {
		if !errors.Is(err, psi.ErrNodeNotFound) {
			logger.Errorw("failed to resolve node", "path", nfw.path, "err", err)
		}

		return nil, 0, syscall.ENOENT
	}

	fh = &psiNodeFileHandle{
		tx:   tx,
		view: nfw,
		node: n,
	}

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

	if err := p.view.prepare(ctx, &p.data, p.node); err != nil {
		return nil, err
	}

	p.materialized = true

	return &p.data, nil
}

var _ fs.NodeOpener = (*psiNodeFileView)(nil)
var _ fs.FileReader = (*psiNodeFileHandle)(nil)
