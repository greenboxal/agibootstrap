package coreapi

import `github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators`

type JournalOperations interface {
	GetHead() (uint64, error)
	Iterate(startIndex uint64, count int) iterators.Iterator[JournalEntry]
	Read(index uint64, dst *JournalEntry) (*JournalEntry, error)
	Write(op *JournalEntry) error
}

type Journal interface {
	JournalOperations

	Close() error
}
