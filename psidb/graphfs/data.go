package graphfs

import "github.com/greenboxal/agibootstrap/pkg/psi"

type NodeFlag int32

const (
	NodeFlagNone NodeFlag = iota
	NodeFlagHasData
	NodeFlagHasDataLink
)

type SerializedNode struct {
	Index   int64    `json:"index"`
	Parent  int64    `json:"parent"`
	Version int64    `json:"version"`
	Path    psi.Path `json:"path"`
	Flags   NodeFlag `json:"flags"`
	Type    string   `json:"type"`
	Data    []byte   `json:"data"`
}

type EdgeFlag int32

const (
	EdgeFlagNone EdgeFlag = iota
	EdgeFlagRegular
	EdgeFlagLink
)

type SerializedEdge struct {
	Index   int64 `json:"index"`
	Version int64 `json:"version"`

	Flags EdgeFlag    `json:"flags"`
	Key   psi.EdgeKey `json:"key"`

	ToIndex int64     `json:"toIndex"`
	ToPath  *psi.Path `json:"toPath,omitempty"`

	Data []byte `json:"data,omitempty"`
}
