package thoughtstream

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type CommHandle struct {
	ID   string
	Name string
	Role msn.Role
}

type Thought struct {
	psi.NodeBase

	Pointer Pointer
	From    CommHandle
	Text    string

	ReplyTo *CommHandle
}

func NewThought() *Thought {
	t := &Thought{}

	t.Init(t, "")

	return t
}

func (t *Thought) PreviousThought() *Thought {
	return t.PreviousSibling().(*Thought)
}

func (t *Thought) NextThought() *Thought {
	return t.NextSibling().(*Thought)
}
