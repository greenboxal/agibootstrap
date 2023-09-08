package query

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Iterator interface {
	Next() bool
	Value() psi.Node
	Len() int
	Err() error
	Close() error
}

func NewIterator(fn func() (psi.Node, bool, error)) Iterator {
	return &funcIterator{fn: fn}
}

func NewChanIterator(ch chan IteratorItem) Iterator {
	return &chanIterator{ch: ch}
}

type IteratorItem struct {
	Value psi.Node
	Err   error
}

type chanIterator struct {
	ch      chan IteratorItem
	current IteratorItem
	done    bool
}

func (c *chanIterator) Next() bool {
	if c.ch == nil {
		return false
	}

	current, ok := <-c.ch

	if !ok {
		c.ch = nil
	}

	c.current = current

	return ok
}

func (c *chanIterator) Value() psi.Node { return c.current.Value }
func (c *chanIterator) Err() error      { return c.current.Err }
func (c *chanIterator) Len() int        { return -1 }

func (c *chanIterator) Close() error {
	close(c.ch)

	return nil
}

type funcIterator struct {
	fn      func() (psi.Node, bool, error)
	current psi.Node
	err     error
}

func (f *funcIterator) Next() bool {
	n, ok, err := f.fn()

	if err != nil {
		f.err = err
		return false
	}

	if !ok {
		return false
	}

	f.current = n

	return true
}

func (f *funcIterator) Value() psi.Node { return f.current }
func (f *funcIterator) Len() int        { return -1 }
func (f *funcIterator) Err() error      { return f.err }
func (f *funcIterator) Close() error    { return nil }

type iteratorWrapper struct {
	iterators.Iterator[psi.Node]
}

func (i iteratorWrapper) Len() int     { return -1 }
func (i iteratorWrapper) Err() error   { return nil }
func (i iteratorWrapper) Close() error { return nil }
