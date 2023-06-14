package vfs

import (
	"strings"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type FsNode interface {
	psi.Node

	Path() string
}

type DirectoryNode struct {
	psi.NodeBase
}

func NewDirectoryNode(path string) *DirectoryNode {
	node := &DirectoryNode{}
	node.SetPath(path)
	return node
}

type FileNode struct {
	psi.NodeBase

	Key  string
	Path string
}

func NewFileNode(path string) *FileNode {
	key := strings.ToLower(path)

	return &FileNode{
		Key:  key,
		Path: path,
	}
}
