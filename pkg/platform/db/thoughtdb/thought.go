package thoughtdb

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type CommHandle struct {
	ID   string
	Name string
	Role msn.Role
}

type Thought struct {
	psi.NodeBase

	Pointer Pointer
	Parents []Pointer

	From CommHandle
	Text string

	ReplyTo *CommHandle
}

var ThoughtType = psi.DefineNodeType[*Thought]()

func NewThought() *Thought {
	t := &Thought{}

	t.Init(t, psi.WithNodeType(ThoughtType))

	return t
}

func (t *Thought) PsiNodeName() string { return t.Pointer.String() }

func (t *Thought) PreviousThought() *Thought {
	return t.PreviousSibling().(*Thought)
}

func (t *Thought) NextThought() *Thought {
	return t.NextSibling().(*Thought)
}

func (t *Thought) Clone() *Thought {
	clone := NewThought()

	clone.Pointer = t.Pointer
	clone.From = t.From
	clone.Text = t.Text
	clone.ReplyTo = t.ReplyTo

	return clone
}
