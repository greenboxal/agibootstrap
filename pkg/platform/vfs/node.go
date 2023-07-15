package vfs

import (
	"context"
	"path"

	"github.com/fsnotify/fsnotify"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

// A Node is a node in the virtual file system.
type Node interface {
	psi.Node

	VfsParent() Node

	GetFileSystem() FileSystem

	GetName() string
	GetPath() string

	Watch() error
	Unwatch() error

	onWatchEvent(ctx context.Context, ev fsnotify.Event) error
}

type NodeBase struct {
	psi.NodeBase

	fs *fileSystem

	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

func (nb *NodeBase) Init(n psi.Node, typ psi.NodeType) {
	nb.NodeBase.Init(n, psi.WithNodeType(typ))

	if nb.fs == nil {
		if p := nb.VfsParent(); p != nil {
			if fs, ok := p.GetFileSystem().(*fileSystem); ok {
				nb.fs = fs
			}
		}
	}

	if nb.fs == nil {
		panic("vfs: node has no file system")
	}
}

func (nb *NodeBase) VfsParent() Node {
	if p, ok := nb.Parent().(Node); ok {
		return p
	}

	return nil
}

func (nb *NodeBase) PsiNodeName() string       { return nb.Name }
func (nb *NodeBase) GetName() string           { return path.Base(nb.Path) }
func (nb *NodeBase) GetPath() string           { return nb.Path }
func (nb *NodeBase) GetFileSystem() FileSystem { return nb.fs }

func (nb *NodeBase) Watch() error   { return nb.fs.addWatch(nb) }
func (nb *NodeBase) Unwatch() error { return nb.fs.removeWatch(nb) }
