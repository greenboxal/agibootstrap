package vfs

import (
	"io/fs"
	"path"
	"path/filepath"
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

type NodeBase struct {
	psi.NodeBase

	fs   FS
	name string
	path string
}

func (nb *NodeBase) PsiNodeName() string { return nb.name }
func (nb *NodeBase) Ast() dst.Node       { return nil }
func (nb *NodeBase) IsContainer() bool   { return true }
func (nb *NodeBase) IsLeaf() bool        { return false }
func (nb *NodeBase) Comments() []string  { return nil }
func (nb *NodeBase) Name() string        { return path.Base(nb.path) }
func (nb *NodeBase) Path() string        { return nb.path }

// A DirectoryNode is a directory in the virtual file system.
type Directory struct {
	NodeBase

	mu sync.RWMutex

	children map[string]FsNode
}

var DirectoryType = psi.DefineNodeType[*Directory](psi.WithRuntimeOnly())

// NewDirectoryNode creates a new DirectoryNode with the specified path.
// The key of the DirectoryNode is set to the lowercase version of the path.
func NewDirectoryNode(fs FS, path string, name string) *Directory {
	dn := &Directory{
		children: map[string]FsNode{},
	}

	if name == "" {
		name = filepath.Base(path)
	}

	dn.fs = fs
	dn.name = name
	dn.path = path

	dn.Init(dn, psi.WithNodeType(DirectoryType))

	return dn
}

func (dn *Directory) Resolve(name string) FsNode {
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
func (dn *Directory) Sync(filterFn func(path string) bool) error {
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
				n = NewDirectoryNode(dn.fs, fullPath, file.Name())
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

type File struct {
	NodeBase
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
