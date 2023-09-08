package journal

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type InMemoryJournal struct {
}

func (j *InMemoryJournal) GetHead() (uint64, error) {
	//TODO implement me
	panic("implement me")
}

func (j *InMemoryJournal) Iterate(startIndex uint64, count int) iterators.Iterator[coreapi.JournalEntry] {
	//TODO implement me
	panic("implement me")
}

func (j *InMemoryJournal) Read(index uint64, dst *coreapi.JournalEntry) (*coreapi.JournalEntry, error) {
	//TODO implement me
	panic("implement me")
}

func (j *InMemoryJournal) Write(op *coreapi.JournalEntry) error {
	//TODO implement me
	panic("implement me")
}

func (j *InMemoryJournal) Close() error {
	//TODO implement me
	panic("implement me")
}
