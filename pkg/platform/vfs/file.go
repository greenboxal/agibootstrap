package vfs

import (
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/multiformats/go-multihash"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type File struct {
	NodeBase

	lastSyncEtag    multihash.Multihash
	currentSyncEtag multihash.Multihash
}

var FileType = psi.DefineNodeType[*File](psi.WithRuntimeOnly())

func NewFileNode(fs FS, path string) *File {
	fn := &File{}

	fn.fs = fs
	fn.name = filepath.Base(path)
	fn.path = path

	fn.Init(fn, psi.WithNodeType(FileType))

	return fn
}

func (f *File) onWatchEvent(ev fsnotify.Event) error {
	if ev.Has(fsnotify.Remove) {
		f.SetParent(nil)
	}

	f.Invalidate()

	return nil
}
