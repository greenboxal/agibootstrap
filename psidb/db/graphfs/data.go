package graphfs

import (
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

	ToIndex int64     `json:"toIndex"`
	ToPath  *psi.Path `json:"toPath,omitempty"`

	Data []byte `json:"data,omitempty"`

	Xmin uint64 `json:"xmin"`
	Xmax uint64 `json:"xmax"`
}
