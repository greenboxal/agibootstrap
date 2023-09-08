package psi

import (
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

type FrozenGraph struct {
	Nodes []*FrozenNode
	Edges []*FrozenEdge
}

type FrozenNode struct {
	Index   int64 `json:"index"`
	Version int64 `json:"version"`

	Path Path   `json:"path"`
	Type string `json:"type"`

	Edges      []cidlink.Link    `json:"edges,omitempty"`
	Attributes map[string]string `json:"attr,omitempty"`
	Children   []EdgeKey         `json:"children,omitempty"`

	Data *cidlink.Link `json:"link,omitempty"`
}

type FrozenEdge struct {
	Key EdgeKey `json:"key"`

	FromIndex int64 `json:"from_index"`
	FromPath  Path  `json:"from_path"`

	ToPath  *Path         `json:"to_path"`
	ToLink  *cidlink.Link `json:"to_link"`
	ToIndex int64         `json:"to_index"`

	Attributes map[string]interface{} `json:"attr,omitempty"`

	Data *cidlink.Link `json:"data,omitempty"`
}
