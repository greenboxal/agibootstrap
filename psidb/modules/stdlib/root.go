package stdlib

import "github.com/greenboxal/agibootstrap/pkg/psi"

type RootNode struct {
	psi.NodeBase

	NodeUUID string `json:"UUID,omitempty"`
}

var RootNodeType = psi.DefineNodeType[*RootNode]()

func (c *RootNode) UUID() string { return c.NodeUUID }
