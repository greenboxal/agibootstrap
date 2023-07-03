package thoughtdb

import (
	"context"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Branch interface {
	psi.Node

	Cursor() Cursor
	Head() *Thought
	Pointer() Pointer

	Commit(ctx context.Context, t *Thought) error

	Fork() Branch
	Merge(ctx context.Context, strategy MergeStrategy, forks ...Branch) error
}

type repoBranch struct {
	psi.NodeBase

	repo *Repo
	head *Thought
}

func newBranch(repo *Repo, head *Thought) *repoBranch {
	b := &repoBranch{
		repo: repo,
		head: head,
	}

	b.Init(b)

	return b
}

func (b *repoBranch) Head() *Thought { return b.head }

func (b *repoBranch) Pointer() Pointer {
	if b.head == nil {
		return Pointer{}
	}

	return b.head.Pointer
}

func (b *repoBranch) Cursor() Cursor {
	c := b.repo.CreateCursor()

	if b.head != nil {
		c.Enqueue(iterators.Single[psi.Node](b.head))
	}

	return c
}

func (b *repoBranch) Commit(ctx context.Context, head *Thought) error {
	if !head.Pointer.IsZero() {
		return errors.New("thought is already committed")
	}

	head.SetParent(b)

	if b.head == nil {
		head.Pointer = Pointer{}
	} else {
		link, err := b.repo.graph.CommitNode(ctx, head)

		if err != nil {
			return err
		}

		head.Pointer = b.head.Pointer.Next(link)
		head.Parents = []Pointer{b.head.Pointer}
	}

	b.head = head
	b.repo.thoughtCache[head.Pointer] = head

	head.Invalidate()

	return head.Update(ctx)
}

func (b *repoBranch) Fork() Branch {
	fork := newBranch(b.repo, b.head)
	fork.SetParent(b)
	return fork
}

func (b *repoBranch) Merge(ctx context.Context, strategy MergeStrategy, forks ...Branch) error {
	return strategy.Merge(ctx, b.repo, b, forks)
}
