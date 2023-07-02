package thoughtdb

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Repo struct {
	psi.NodeBase

	graph *graphstore.IndexedGraph

	thoughtCache map[Pointer]*Thought
}

func NewRepo(graph *graphstore.IndexedGraph) *Repo {
	r := &Repo{
		graph:        graph,
		thoughtCache: map[Pointer]*Thought{},
	}

	r.Init(r, "<tdb-repo>")

	return r
}

func (r *Repo) PsiNodeName() string             { return "LogManager" }
func (r *Repo) Graph() *graphstore.IndexedGraph { return r.graph }

func (r *Repo) CreateBranch() Branch {
	b := newBranch(r, nil)
	b.SetParent(r)
	return b
}

func (r *Repo) CreateCursor() Cursor {
	return &repoCursor{repo: r, Cursor: psi.NewCursor()}
}

func (r *Repo) ResolvePointer(pointer Pointer) (thought *Thought, err error) {
	return r.thoughtCache[pointer], nil
}

func (r *Repo) Checkout(head *Thought) (branch Branch, err error) {
	return newBranch(r, head), nil
}
