package vfs

import (
	"io/fs"
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

	Path() string
}

// A DirectoryNode is a directory in the virtual file system.
type DirectoryNode struct {
	psi.NodeBase

	fs   FS
	key  string
	path string

	mu       sync.RWMutex
	children map[string]FsNode
}

// NewDirectoryNode creates a new DirectoryNode with the specified path.
// The key of the DirectoryNode is set to the lowercase version of the path.
func NewDirectoryNode(fs FS, path string) *DirectoryNode {
	key := strings.ToLower(path)

	dn := &DirectoryNode{
		fs:   fs,
		key:  key,
		path: path,
	}

	return dn
}

func (dn *DirectoryNode) Ast() dst.Node      { return nil }
func (dn *DirectoryNode) IsContainer() bool  { return true }
func (dn *DirectoryNode) IsLeaf() bool       { return false }
func (dn *DirectoryNode) Comments() []string { return nil }
func (dn *DirectoryNode) Path() string       { return dn.path }

// Sync synchronizes the DirectoryNode with the underlying filesystem.
// It scans the directory and updates the children nodes to reflect the current state of the filesystem.
// Any nodes that no longer exist in the filesystem are removed.
func (dn *DirectoryNode) Sync(recursive bool) error {
	files, err := fs.ReadDir(dn.fs, dn.path)
	if err != nil {
		return err
	}

	dn.mu.Lock()
	defer dn.mu.Unlock()

	// Remove nodes that no longer exist in the filesystem
	for _, file := range files {
		filePath := filepath.Join(dn.path, file.Name())
		if file.IsDir() {
			dirNode := NewDirectoryNode(dn.fs, filePath)
			dirNode.SetParent(dn)
			dn.children[filePath] = dirNode

			if recursive {
				dirNode.Sync(recursive)
			}
		} else {
			fileNode := NewFileNode(dn.fs, filePath)
			fileNode.SetParent(dn)
			dn.children[filePath] = fileNode
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
func (fn *FileNode) Path() string       { return fn.path }

func NewFileNode(fs FS, path string) *FileNode {
	key := strings.ToLower(path)

	fn := &FileNode{
		fs:   fs,
		key:  key,
		path: path,
	}

	fn.Initialize(fn, key)

	return fn
}
