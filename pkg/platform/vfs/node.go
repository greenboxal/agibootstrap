package vfs

import (
	"path"

	"github.com/fsnotify/fsnotify"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

// A Node is a node in the virtual file system.
type Node interface {
	psi.Node

	Name() string
	Path() string

	onWatchEvent(ev fsnotify.Event) error
}

type NodeBase struct {
	psi.NodeBase

	fs   FS
	name string
	path string
}

func (nb *NodeBase) PsiNodeName() string { return nb.name }
func (nb *NodeBase) Name() string        { return path.Base(nb.path) }
func (nb *NodeBase) Path() string        { return nb.path }
