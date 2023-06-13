package vfs

import (
	"strings"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

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
