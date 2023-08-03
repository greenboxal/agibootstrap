package graphfs

import (
	"fmt"
	"os"
	"sync"

	"github.com/greenboxal/aip/aip-forddb/pkg/typesystem"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/tidwall/wal"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Journal struct {
	logger *zap.SugaredLogger

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

func (j *Journal) Iterate(startIndex uint64, count int) iterators.Iterator[JournalEntry] {
	index := startIndex

	return iterators.NewIterator(func() (res JournalEntry, ok bool) {
		j.mu.RLock()
		defer j.mu.RUnlock()

		if index >= j.nextIndex || (count >= 0 && count >= int(j.nextIndex-index)) {
			return JournalEntry{}, false
		}

		_, err := j.Read(index, &res)

		if err == wal.ErrNotFound {
			return JournalEntry{}, false
		}

		if err != nil {
			j.logger.Error(err)
			return JournalEntry{}, false
		}

		index++

		return res, true
	})
}

func (j *Journal) Read(index uint64, dst *JournalEntry) (*JournalEntry, error) {
	data, err := j.wal.Read(index)

	if err != nil {
		return nil, err
	}

	wrapped, err := ipld.DecodeUsingPrototype(data, dagcbor.Decode, JournalEntryType.IpldPrototype())

	if err != nil {
		return nil, err
	}

	entry, ok := typesystem.TryUnwrap[*JournalEntry](wrapped)

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

func (j *Journal) Write(op JournalEntry) (JournalEntry, error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	idx := j.nextIndex

	if op.Op == JournalOpBegin && op.Xid == 0 {
		op.Xid = idx
	}

	data, err := ipld.Encode(typesystem.Wrap(op), dagcbor.Encode)

	if err != nil {
		return op, err
	}

	if err := j.wal.Write(idx, data); err != nil {
		return op, err
	}

	j.nextIndex++

	return op, nil
}

func (j *Journal) Close() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.wal.Close()
}

type JournalEntry struct {
	Ts    int64           `json:"ts"`
	Op    JournalOp       `json:"op"`
	Xid   uint64          `json:"xid"`
	Inode int64           `json:"inode"`
	Path  *psi.Path       `json:"path,omitempty"`
	Node  *SerializedNode `json:"node,omitempty"`
	Edge  *SerializedEdge `json:"edge,omitempty"`
}

var JournalEntryType = typesystem.TypeOf((*JournalEntry)(nil))

type JournalOp int

const (
	JournalOpInvalid JournalOp = iota
	JournalOpBegin
	JournalOpCommit
	JournalOpRollback
	JournalOpWrite
	JournalOpSetEdge
	JournalOpRemoveEdge
)

func (op JournalOp) String() string {
	switch op {
	case JournalOpInvalid:
		return "invalid"
	case JournalOpBegin:
		return "begin"
	case JournalOpCommit:
		return "commit"
	case JournalOpRollback:
		return "rollback"
	case JournalOpWrite:
		return "write"
	case JournalOpSetEdge:
		return "set_edge"
	case JournalOpRemoveEdge:
		return "remove_edge"
	default:
		return "unknown"
	}
}
