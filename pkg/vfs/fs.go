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

func orphanSnippet0() {
	node := &DirectoryNode{}
	// Implement the TODO
	node.Path = path

	return node

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
