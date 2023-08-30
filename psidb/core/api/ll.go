package coreapi

import "sync"

type LinkedList[T any] struct {
	sync.RWMutex
	head *ListHead[T]
	tail *ListHead[T]
}

func (l *LinkedList[T]) Head() *ListHead[T] {
	l.RLock()
	defer l.RUnlock()

	return l.head
}

func (l *LinkedList[T]) Tail() *ListHead[T] {
	l.RLock()
	defer l.RUnlock()

	return l.tail
}

func (l *LinkedList[T]) Add(n *ListHead[T]) *ListHead[T] {
	l.Lock()
	defer l.Unlock()

	return l.AddUnlocked(n)
}

func (l *LinkedList[T]) AddUnlocked(n *ListHead[T]) *ListHead[T] {
	if l.head == nil {
		l.head = n
		l.tail = l.head
		return n
	}

	l.tail.Next = n
	l.tail = l.tail.Next

	return n
}

func (l *LinkedList[T]) Remove(n *ListHead[T]) {
	if n == nil {
		return
	}

	l.Lock()
	defer l.Unlock()

	l.RemoveUnlocked(n)
}

func (l *LinkedList[T]) RemoveUnlocked(n *ListHead[T]) {
	if l.head == n {
		l.head = n.Next
	}

	if l.tail == n {
		l.tail = n.Prev
	}

	n.Remove()
}

type ListHead[T any] struct {
	sync.Mutex

	Next *ListHead[T]
	Prev *ListHead[T]

	Value T
}

func (lh *ListHead[T]) AddNext(next *ListHead[T]) {
	if next.Prev != nil || next.Next != nil {
		panic("Next is already in a list")
	}

	lh.Lock()
	defer lh.Unlock()

	lh.AddNextUnlocked(next)
}

func (lh *ListHead[T]) AddNextUnlocked(next *ListHead[T]) {
	next.Next = lh.Next
	lh.Next = next
	next.Prev = lh
}

func (lh *ListHead[T]) AddPrevious(prev *ListHead[T]) {
	if prev.Prev != nil || prev.Next != nil {
		panic("Prev is already in a list")
	}

	lh.Lock()
	defer lh.Unlock()

	lh.AddPreviousUnlocked(prev)
}

func (lh *ListHead[T]) AddPreviousUnlocked(prev *ListHead[T]) {
	prev.Prev = lh.Prev
	lh.Prev = prev
	prev.Next = lh
}

func (lh *ListHead[T]) Remove() {
	lh.Lock()
	defer lh.Unlock()

	lh.RemoveUnlocked()
}

func (lh *ListHead[T]) RemoveUnlocked() {
	if lh.Prev != nil {
		lh.Prev.Next = lh.Next
	}

	if lh.Next != nil {
		lh.Next.Prev = lh.Prev
	}
}
