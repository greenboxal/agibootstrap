package graphfs

import (
	"fmt"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type NodeFlag int32

const (
	NodeFlagNone NodeFlag = iota
	NodeFlagHasData
	NodeFlagHasDataLink

	NodeFlagRemoved NodeFlag = 1 << 30
)

type SerializedNode struct {
	Index   int64         `json:"index"`
	Parent  int64         `json:"parent"`
	Version int64         `json:"version"`
	Path    psi.Path      `json:"path"`
	Flags   NodeFlag      `json:"flags"`
	Type    string        `json:"type"`
	Data    []byte        `json:"data"`
	Link    *cidlink.Link `json:"link"`
}

func (n SerializedNode) IsRemoved() bool { return n.Flags&NodeFlagRemoved != 0 }

func (n SerializedNode) String() string {
	return fmt.Sprintf(
		"Node{Index: %d, Parent: %d, Version: %d, Path: %s, Flags: %d, Type: %s, Data: %s, Link: %s}",
		n.Index, n.Parent, n.Version, n.Path, n.Flags, n.Type, n.Data, n.Link,
	)
}

type EdgeFlag int32

const (
	EdgeFlagNone EdgeFlag = iota

	EdgeFlagRegular
	EdgeFlagLink
	EdgeFlagModes = EdgeFlagRegular | EdgeFlagLink

	EdgeFlagRemoved EdgeFlag = 1 << 30
)

type SerializedEdge struct {
	Index   int64 `json:"index"`
	Version int64 `json:"version"`

	Flags EdgeFlag    `json:"flags"`
	Key   psi.EdgeKey `json:"key"`

	ToIndex int64    `json:"toIndex"`
	ToPath  psi.Path `json:"toPath,omitempty"`

	Data []byte `json:"data,omitempty"`
}

func (e SerializedEdge) IsRemoved() bool { return e.Flags&EdgeFlagRemoved != 0 }

func (e SerializedEdge) String() string {
	return fmt.Sprintf(
		"Edge{Index: %d, Version: %d, Flags: %d, Key: %s, ToIndex: %d, ToPath: %s, Data: %s}",
		e.Index, e.Version, e.Flags, e.Key, e.ToIndex, e.ToPath, e.Data,
	)
}
