package vfs

import (
	"context"
	"io/fs"
	"path"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

// Directory is a directory in the virtual file system.
type Directory struct {
	NodeBase `ipld:",inline"`

	mu sync.RWMutex

	vfsm *Manager `inject:""`
}

var DirectoryType = psi.DefineNodeType[*Directory](psi.WithRuntimeOnly())

// NewDirectoryNode creates a new DirectoryNode with the specified path.
// The key of the DirectoryNode is set to the lowercase version of the path.
func newDirectoryNode(fs *fileSystem, path string, name string) *Directory {
	dn := &Directory{}

	if name == "" {
		name = filepath.Base(path)
	}

	dn.fs = fs
	dn.Name = name
	dn.Path = path

	dn.Init(dn)

	return dn
}

func (dn *Directory) Init(self psi.Node) {
	dn.NodeBase.Init(self, DirectoryType)
}

func (dn *Directory) Lookup(name string) (Node, error) {
	dn.mu.RLock()
	defer dn.mu.RUnlock()

	for it := dn.ChildrenIterator(); it.Next(); {
		child, ok := it.Value().(Node)

		if !ok {
			continue
		}

		if child.GetName() == name {
			return child, nil
		}
	}

	return nil, psi.ErrNodeNotFound
}

// Sync synchronizes the DirectoryNode with the underlying filesystem.
// It scans the directory and updates the children nodes to reflect the current state of the filesystem.
// Any nodes that no longer exist in the filesystem are removed.
func (dn *Directory) Sync(filterFn func(path string) bool) error {
	dn.mu.Lock()
	defer dn.mu.Unlock()

	files, err := fs.ReadDir(dn.fs, dn.Path)

	if err != nil {
		return err
	}

	changes := make(map[string]Node)

	for _, file := range files {
		var n Node

		fullPath := path.Join(dn.Path, file.Name())

		if filterFn != nil && !filterFn(fullPath) {
			continue
		}

		for it := dn.ChildrenIterator(); it.Next(); {
			child, ok := it.Value().(Node)

			if !ok {
				continue
			}

			if child.GetPath() == fullPath {
				n = child
				break
			}
		}

		if n == nil {
			if file.IsDir() {
				n = newDirectoryNode(dn.fs, fullPath, file.Name())
			} else {
				n = newFileNode(dn.fs, fullPath)
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

		if _, ok := changes[child.GetName()]; !ok {
			child.SetParent(nil)

			continue
		}
	}

	return nil
}

func (dn *Directory) onWatchEvent(ctx context.Context, ev fsnotify.Event) error {
	if ev.Has(fsnotify.Remove) {
		dn.SetParent(nil)
	} else {
		if err := dn.Sync(nil); err != nil {
			return err
		}
	}

	return dn.Update(ctx)
}

func (dn *Directory) GetOrCreateFile(ctx context.Context, name string) (*File, error) {
	existing, err := dn.Lookup(name)

	if err == nil {
		return existing.(*File), nil
	}

	if err != psi.ErrNodeNotFound {
		return nil, err
	}

	f := newFileNode(dn.fs, path.Join(dn.Path, name))
	f.SetParent(dn)

	dn.fs.nodeMap[f.Path] = f

	return f, nil
}
