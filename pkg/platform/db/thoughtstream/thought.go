package thoughtstream

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Thought interface {
	psi.Node

	PreviousThought() Thought
	NextThought() Thought
}

type ThoughtBase struct {
	psi.NodeBase
}

func (t *ThoughtBase) PreviousThought() Thought {
	return t.PreviousSibling().(Thought)
}

func (t *ThoughtBase) NextThought() Thought {
	return t.NextSibling().(Thought)
}
