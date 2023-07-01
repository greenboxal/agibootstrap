package thoughtdb

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Cursor interface {
	psi.Cursor

	Thought() *Thought
	Pointer() Pointer

	PushThought(head *Thought)
	PushPointer(pointer Pointer) error

	IterateParents() iterators.Iterator[*Thought]

	Clone() Cursor
}

type repoCursor struct {
	psi.Cursor

	repo *Repo
}

func (r *repoCursor) Pointer() Pointer  { return r.Thought().Pointer }
func (r *repoCursor) Thought() *Thought { return r.Value().(*Thought) }

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
	for r.Cursor.Next() {
		_, ok := r.Value().(*Thought)

		if !ok {
			continue
		}

		return true
	}

	return false
}

func (r *repoCursor) Parents() iterators.Iterator[*Thought] {
	return iterators.Map(iterators.FromSlice(r.Thought().Parents), func(p Pointer) *Thought {
		t, err := r.repo.ResolvePointer(p)

		if err != nil {
			panic(err)
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
	cur.SetCurrent(r.Thought())
	return &repoCursor{Cursor: cur, repo: r.repo}
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
	it.cursor.EnqueueParents()

	return true
}
