package vfs

import (
	"strings"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

// A FsNode is a node in the virtual file system.
type FsNode interface {
	psi.Node

	Path() string
}

// A DirectoryNode is a directory in the virtual file system.
type DirectoryNode struct {
	psi.NodeBase

	key  string
	path string
}

// NewDirectoryNode creates a new DirectoryNode with the specified path.
// The key of the DirectoryNode is set to the lowercase version of the path.
func NewDirectoryNode(path string) *DirectoryNode {
	key := strings.ToLower(path)

	dn := &DirectoryNode{
		key:  key,
		path: path,
	}

	return dn
}

type FileNode struct {
	psi.NodeBase

	Key  string
	Path string
}

func NewFileNode(path string) *FileNode {
	key := strings.ToLower(path)

	fn := &FileNode{
		Key:  key,
		Path: path,
	}

	return fn
}
