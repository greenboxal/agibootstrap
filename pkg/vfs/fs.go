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

func (dn *DirectoryNode) Sync() error {
	files, err := fs.ReadDir(dn.fs, dn.path)

	if err != nil {
		return err
	}

	dn.mu.Lock()
	defer dn.mu.Unlock()

	// Remove nodes that no longer exist in the filesystem
	for _, file := range files {
		node, ok := dn.children[file.Name()]
		if ok {
			if file.IsDir() && _, ok := node.(*DirectoryNode); !ok {
				delete(dn.children, file.Name())
			} else if !file.IsDir() && _, ok := node.(*FileNode); !ok {
				delete(dn.children, file.Name())
			}
		} else {
			filePath := filepath.Join(dn.path, file.Name())
			if file.IsDir() {
				dirNode := NewDirectoryNode(dn.fs, filePath)
				dirNode.SetParent(dn)
				dn.children[filePath] = dirNode
			} else {
				fileNode := NewFileNode(dn.fs, filePath)
				fileNode.SetParent(dn)
				dn.children[filePath] = fileNode
			}
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
