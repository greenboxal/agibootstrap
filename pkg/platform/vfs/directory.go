package vfs

import (
	"io/fs"
	"path"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

// Directory is a directory in the virtual file system.
type Directory struct {
	NodeBase

	mu sync.RWMutex
}

var DirectoryType = psi.DefineNodeType[*Directory](psi.WithRuntimeOnly())

// NewDirectoryNode creates a new DirectoryNode with the specified path.
// The key of the DirectoryNode is set to the lowercase version of the path.
func NewDirectoryNode(fs FS, path string, name string) *Directory {
	dn := &Directory{}

	if name == "" {
		name = filepath.Base(path)
	}

	dn.fs = fs
	dn.name = name
	dn.path = path

	dn.Init(dn, psi.WithNodeType(DirectoryType))

	return dn
}

// Sync synchronizes the DirectoryNode with the underlying filesystem.
// It scans the directory and updates the children nodes to reflect the current state of the filesystem.
// Any nodes that no longer exist in the filesystem are removed.
func (dn *Directory) Sync(filterFn func(path string) bool) error {
	dn.mu.Lock()
	defer dn.mu.Unlock()

	files, err := fs.ReadDir(dn.fs, dn.path)

	if err != nil {
		return err
	}

	changes := make(map[string]Node)

	for _, file := range files {
		var n Node

		fullPath := path.Join(dn.path, file.Name())

		if filterFn != nil && !filterFn(fullPath) {
			continue
		}

		for it := dn.ChildrenIterator(); it.Next(); {
			child, ok := it.Value().(Node)

			if !ok {
				continue
			}

			if child.Path() == fullPath {
				n = child
				break
			}
		}

		if n == nil {
			if file.IsDir() {
				n = NewDirectoryNode(dn.fs, fullPath, file.Name())
			} else {
				n = NewFileNode(dn.fs, fullPath)
			}

			n.SetParent(dn)
		}

		changes[file.Name()] = n
	}

	for it := dn.ChildrenIterator(); it.Next(); {
		child, ok := it.Value().(Node)

		if !ok {
			continue
		}

		if _, ok := changes[child.Name()]; !ok {
			child.SetParent(nil)

			continue
		}
	}

	return nil
}

func (dn *Directory) onWatchEvent(ev fsnotify.Event) error {
	if ev.Has(fsnotify.Remove) {
		dn.SetParent(nil)
	}

	dn.Invalidate()

	return nil
}
