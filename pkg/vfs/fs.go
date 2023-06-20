package vfs

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dave/dst"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type FS interface {
	fs.FS
}

// A FsNode is a node in the virtual file system.
type FsNode interface {
	psi.Node

	Name() string
	Path() string
}

// A DirectoryNode is a directory in the virtual file system.
type DirectoryNode struct {
	psi.NodeBase

	fs   FS
	key  string
	path string

	mu sync.RWMutex

	children map[string]FsNode
}

// NewDirectoryNode creates a new DirectoryNode with the specified path.
// The key of the DirectoryNode is set to the lowercase version of the path.
func NewDirectoryNode(fs FS, path string) *DirectoryNode {
	key := strings.ToLower(path)

	dn := &DirectoryNode{
		fs:       fs,
		key:      key,
		path:     path,
		children: map[string]FsNode{},
	}

	dn.Init(dn, path)

	return dn
}

func (dn *DirectoryNode) Ast() dst.Node      { return nil }
func (dn *DirectoryNode) IsContainer() bool  { return true }
func (dn *DirectoryNode) IsLeaf() bool       { return false }
func (dn *DirectoryNode) Comments() []string { return nil }
func (dn *DirectoryNode) Name() string       { return filepath.Base(dn.path) }
func (dn *DirectoryNode) Path() string       { return dn.path }

func (dn *DirectoryNode) Resolve(name string) FsNode {
	dn.mu.RLock()
	defer dn.mu.RUnlock()

	if child, ok := dn.children[name]; ok {
		return child
	}

	return nil
}

// Sync synchronizes the DirectoryNode with the underlying filesystem.
// It scans the directory and updates the children nodes to reflect the current state of the filesystem.
// Any nodes that no longer exist in the filesystem are removed.
func (dn *DirectoryNode) Sync(filterFn func(path string) bool) error {
	dn.mu.Lock()
	defer dn.mu.Unlock()

	files, err := fs.ReadDir(dn.fs, dn.path)

	if err != nil {
		return err
	}

	changes := make(map[string]FsNode)

	for _, file := range files {
		fullPath := path.Join(dn.path, file.Name())

		if filterFn != nil && !filterFn(fullPath) {
			continue
		}

		n := dn.children[file.Name()]

		if n == nil {
			if file.IsDir() {
				n = NewDirectoryNode(dn.fs, fullPath)
			} else {
				n = NewFileNode(dn.fs, fullPath)
			}

			n.SetParent(dn)

			dn.children[file.Name()] = n
		}

		changes[file.Name()] = n
	}

	for _, child := range dn.children {
		if _, ok := changes[child.Name()]; !ok {
			child.SetParent(nil)

			delete(dn.children, child.Name())

			continue
		}
	}

	return nil
}

type FileNode struct {
	psi.NodeBase

	fs   FS
	key  string
	path string
}

func (fn *FileNode) Ast() dst.Node      { return nil }
func (fn *FileNode) IsContainer() bool  { return true }
func (fn *FileNode) IsLeaf() bool       { return false }
func (fn *FileNode) Comments() []string { return nil }
func (fn *FileNode) Name() string       { return filepath.Base(fn.path) }
func (fn *FileNode) Path() string       { return fn.path }

func NewFileNode(fs FS, path string) *FileNode {
	fn := &FileNode{
		fs:   fs,
		path: path,
	}

	fn.Init(fn, path)

	return fn
}
