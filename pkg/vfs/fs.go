package vfs

import (
	"io/fs"
	"path/filepath"
	"strings"

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

	for _, file := range files {
		name := file.Name()
		if !file.IsDir() {
			// Skip directories
			continue
		}

		// Add a new child file node to the directory node
		filePath := filepath.Join(dn.path, name)
		fileNode := NewFileNode(dn.fs, filePath)
		dn.addChildNode(fileNode)
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
