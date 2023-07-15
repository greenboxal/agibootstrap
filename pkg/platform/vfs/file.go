package vfs

import (
	"context"
	"io"
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ipld/go-ipld-prime"

	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type File struct {
	NodeBase `ipld:",inline"`

	mu sync.Mutex

	lastStat   fs.FileInfo
	lastSyncAt time.Time

	isCached        bool
	cachedTimestamp time.Time
	cachedLink      ipld.Link

	openRefs int
}

var FileType = psi.DefineNodeType[*File](psi.WithRuntimeOnly())

func newFileNode(fs *fileSystem, path string) *File {
	fn := &File{}

	fn.fs = fs
	fn.Name = filepath.Base(path)
	fn.Path = path

	fn.Init(fn)

	return fn
}

func (f *File) Init(self psi.Node) {
	f.NodeBase.Init(self, FileType)
}

func (f *File) Open() (repofs.FileHandle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.openRefs++

	if !f.isCached {
		f.isCached = true
		f.updateCache()
	}

	fh := &fileHandle{file: f}

	return fh, nil
}

func (f *File) Sync() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	stat, err := fs.Stat(f.fs, f.Path)

	if err != nil {
		return err
	}

	previousStat := f.lastStat
	f.lastStat = stat
	f.lastSyncAt = time.Now()

	if stat.ModTime() != previousStat.ModTime() {
		f.updateCache()
	}

	return nil
}

func (f *File) onWatchEvent(ctx context.Context, ev fsnotify.Event) error {
	if ev.Has(fsnotify.Remove) {
		f.SetParent(nil)
	} else {
		if err := f.Sync(); err != nil {
			return err
		}
	}

	return f.Update(ctx)
}

func (f *File) notifyClose(fh *fileHandle) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.openRefs--

	if f.openRefs == 0 {
		f.isCached = false
	}
}

func (f *File) updateCache() {
	if !f.isCached {
		return
	}
}

type fileHandle struct {
	mu     sync.Mutex
	file   *File
	closed bool
}

func (fh *fileHandle) Get() (io.ReadCloser, error) {
	return fh.file.fs.Open(fh.file.Path)
}

func (fh *fileHandle) Put(src io.Reader) error {
	return fh.file.fs.WriteFile(fh.file.Path, src)
}

func (fh *fileHandle) Close() error {
	fh.mu.Lock()
	defer fh.mu.Unlock()

	if fh.closed {
		return nil
	}

	fh.file.notifyClose(fh)

	fh.closed = true

	return nil
}
