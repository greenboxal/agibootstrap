package thoughtstream

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Cursor interface {
	psi.Cursor

	Thought() *Thought
}

type Stream interface {
	Cursor

	Append(t *Thought) error
}

type cursor struct {
	psi.Cursor
}

func (s *cursor) Thought() *Thought {
	return s.Node().(*Thought)
}

type stream struct {
	cursor
}

func (s *stream) Append(t *Thought) error {
	s.InsertAfter(t)
	s.SetNext(t)

	return nil
}
