package thoughtdb

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Cursor interface {
	psi.Cursor

	Thought() *Thought
	Pointer() Pointer

	PushThought(head *Thought)
	PushPointer(pointer Pointer) error

	PushParents()
	EnqueueParents()

	IterateParents() iterators.Iterator[*Thought]

	Clone() Cursor
}

type repoCursor struct {
	psi.Cursor

	repo *Repo

	current *Thought
	pointer Pointer
}

func (r *repoCursor) Pointer() Pointer  { return r.pointer }
func (r *repoCursor) Thought() *Thought { return r.current }

func (r *repoCursor) PushThought(t *Thought) {
	r.Enqueue(iterators.Single[psi.Node](t))
}

func (r *repoCursor) PushPointer(p Pointer) error {
	t, err := r.repo.ResolvePointer(p)

	if err != nil {
		return err
	}

	r.PushThought(t)

	return nil
}

func (r *repoCursor) Next() bool {
	if !r.Cursor.Next() {
		return false
	}

	if v, ok := r.Cursor.Value().(*Thought); ok && v != nil {
		r.current = v
		r.pointer = r.current.Pointer
	} else {
		r.current = nil
		r.pointer = Pointer{}
	}

	return true
}

func (r *repoCursor) Parents() iterators.Iterator[*Thought] {
	t := r.Thought()

	if t == nil {
		return iterators.Empty[*Thought]()
	}

	parents := iterators.Filter(iterators.FromSlice(t.Parents), func(t Pointer) bool {
		return !t.IsZero() && !t.IsRoot()
	})

	return iterators.Map(parents, func(p Pointer) *Thought {
		t, err := r.repo.ResolvePointer(p)

		if err != nil {
			panic(err)
		}

		if t == nil {
			panic("nil thought")
		}

		return t
	})
}

func (r *repoCursor) EnqueueParents() {
	r.Enqueue(iterators.FilterIsInstance[*Thought, psi.Node](r.Parents()))
}

func (r *repoCursor) PushParents() {
	r.Push(iterators.FilterIsInstance[*Thought, psi.Node](r.Parents()))
}

func (r *repoCursor) IterateParents() iterators.Iterator[*Thought] {
	return &parentsIterator{cursor: r}
}

func (r *repoCursor) Clone() Cursor {
	cur := psi.NewCursor()
	clone := &repoCursor{Cursor: cur, repo: r.repo}

	clone.PushThought(r.Thought())

	return clone
}

type parentsIterator struct {
	cursor  *repoCursor
	current *Thought
}

func (it *parentsIterator) Value() *Thought { return it.current }

func (it *parentsIterator) Next() bool {
	for !it.cursor.Next() {
		if !it.cursor.Pop() {
			return false
		}
	}

	it.current = it.cursor.Thought()
	it.cursor.PushParents()

	return true
}
