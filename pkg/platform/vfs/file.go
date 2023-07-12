package vfs

import (
	"context"
	"io"
	"io/fs"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type File struct {
	NodeBase

	mu sync.Mutex

	lastStat   fs.FileInfo
	lastSyncAt time.Time

	isCached bool
	openRefs int
}

var FileType = psi.DefineNodeType[*File](psi.WithRuntimeOnly())

func newFileNode(fs *fileSystem, path string) *File {
	fn := &File{}

	fn.fs = fs
	fn.name = filepath.Base(path)
	fn.path = path

	fn.Init(fn, psi.WithNodeType(FileType))

	return fn
}

func (f *File) Open() (repofs.FileHandle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.openRefs++

	fh := &fileHandle{file: f}

	return fh, nil
}

func (f *File) Sync() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	stat, err := fs.Stat(f.fs, f.path)

	if err != nil {
		return err
	}

	if f.lastStat.ModTime() != stat.ModTime() {
		f.onChanged()
	}

	f.lastStat = stat
	f.lastSyncAt = time.Now()

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

func (f *File) onChanged() {
}

type fileHandle struct {
	mu     sync.Mutex
	file   *File
	closed bool
}

func (fh *fileHandle) Get() (io.ReadCloser, error) {
	return fh.file.fs.Open(fh.file.path)
}

func (fh *fileHandle) Put(src io.Reader) error {
	return fh.file.fs.WriteFile(fh.file.path, src)
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
