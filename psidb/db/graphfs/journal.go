package graphfs

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/tidwall/wal"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"
	"github.com/greenboxal/agibootstrap/psidb/core/api"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type Journal struct {
	logger *otelzap.SugaredLogger

	mu  sync.RWMutex
	wal *wal.Log

	nextIndex uint64
}

func OpenJournal(path string) (*Journal, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create journal directory: %s", path)
	}

	j := &Journal{
		logger: logging.GetLogger("graphfs/journal"),
	}

	opts := *wal.DefaultOptions

	w, err := wal.Open(path, &opts)

	if err != nil {
		return nil, err
	}

	last, err := w.LastIndex()

	if err != nil {
		return nil, err
	}

	j.wal = w
	j.nextIndex = last + 1

	return j, nil
}

func (j *Journal) Iterate(startIndex uint64, count int) iterators.Iterator[coreapi.JournalEntry] {
	index := startIndex
	total := 0

	return iterators.NewIterator(func() (res coreapi.JournalEntry, ok bool) {
		j.mu.RLock()
		defer j.mu.RUnlock()

		if index >= j.nextIndex || total >= count {
			return coreapi.JournalEntry{}, false
		}

		_, err := j.Read(index, &res)

		if err == wal.ErrNotFound {
			return coreapi.JournalEntry{}, false
		}

		if err != nil {
			if err != wal.ErrNotFound && err != io.EOF {
				j.logger.Error(err)
			}

			return coreapi.JournalEntry{}, false
		}

		index++
		total++

		return res, true
	})
}

func (j *Journal) Read(index uint64, dst *coreapi.JournalEntry) (*coreapi.JournalEntry, error) {
	data, err := j.wal.Read(index)

	if err != nil {
		if err == wal.ErrNotFound {
			return nil, io.EOF
		}

		return nil, err
	}

	wrapped, err := ipld.DecodeUsingPrototype(data, dagcbor.Decode, coreapi.JournalEntryType.IpldPrototype())

	if err != nil {
		return nil, err
	}

	entry, ok := typesystem.TryUnwrap[*coreapi.JournalEntry](wrapped)

	if !ok {
		return nil, fmt.Errorf("unexpected type %T", wrapped)
	}

	if dst != nil {
		*dst = *entry
	} else {
		dst = entry
	}

	return dst, nil
}

func (j *Journal) Write(op *coreapi.JournalEntry) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	idx := j.nextIndex

	if op.Xid == 0 {
		if op.Op == coreapi.JournalOpBegin {
			op.Xid = idx
		} else {
			panic("invalid journal entry")
		}
	}

	op.Rid = idx

	data, err := ipld.Encode(typesystem.Wrap(op), dagcbor.Encode)

	if err != nil {
		return err
	}

	if err := j.wal.Write(idx, data); err != nil {
		return err
	}

	j.nextIndex++

	return nil
}

func (j *Journal) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.wal.Close()
}

func (j *Journal) GetHead() (uint64, error) {
	return j.wal.LastIndex()
}
