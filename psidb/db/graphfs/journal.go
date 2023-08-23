package graphfs

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/tidwall/wal"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/typesystem"

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
	total := 0

	return iterators.NewIterator(func() (res JournalEntry, ok bool) {
		j.mu.RLock()
		defer j.mu.RUnlock()

		if index >= j.nextIndex || total >= count {
			return JournalEntry{}, false
		}

		_, err := j.Read(index, &res)

		if err == wal.ErrNotFound {
			return JournalEntry{}, false
		}

		if err != nil {
			if err != wal.ErrNotFound && err != io.EOF {
				j.logger.Error(err)
			}

			return JournalEntry{}, false
		}

		index++
		total++

		return res, true
	})
}

func (j *Journal) Read(index uint64, dst *JournalEntry) (*JournalEntry, error) {
	data, err := j.wal.Read(index)

	if err != nil {
		if err == wal.ErrNotFound {
			return nil, io.EOF
		}

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

func (j *Journal) Write(op *JournalEntry) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	idx := j.nextIndex

	if op.Xid == 0 {
		if op.Op == JournalOpBegin {
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

type JournalEntry struct {
	Ts  int64     `json:"ts"`
	Op  JournalOp `json:"op"`
	Rid uint64    `json:"rid"`
	Xid uint64    `json:"xid"`

	Inode int64     `json:"inode"`
	Path  *psi.Path `json:"path,omitempty"`

	Node *SerializedNode `json:"node,omitempty"`
	Edge *SerializedEdge `json:"edge,omitempty"`

	Confirmation *psi.Confirmation `json:"confirmation,omitempty"`
	Notification *psi.Notification `json:"notification,omitempty"`

	Promises []psi.Promise `json:"promises,omitempty"`
}

func (e JournalEntry) String() string {
	return fmt.Sprintf("JournalEntry{ts=%d, op=%s, rid=%d, xid=%d, inode=%d, path=%s, node=%s, edge=%s}",
		e.Ts, e.Op, e.Rid, e.Xid, e.Inode, e.Path, e.Node, e.Edge)
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

	JournalOpNotify
	JournalOpConfirm

	JournalOpWait
	JournalOpSignal
)

func (op JournalOp) String() string {
	switch op {
	case JournalOpInvalid:
		return "JOURNAL_OP_INVALID"
	case JournalOpBegin:
		return "JOURNAL_OP_BEGIN"
	case JournalOpCommit:
		return "JOURNAL_OP_COMMIT"
	case JournalOpRollback:
		return "JOURNAL_OP_ROLLBACK"
	case JournalOpWrite:
		return "JOURNAL_OP_WRITE"
	case JournalOpSetEdge:
		return "JOURNAL_OP_SET_EDGE"
	case JournalOpRemoveEdge:
		return "JOURNAL_OP_REMOVE_EDGE"
	case JournalOpNotify:
		return "JOURNAL_OP_NOTIFY"
	case JournalOpConfirm:
		return "JOURNAL_OP_CONFIRM"
	case JournalOpWait:
		return "JOURNAL_OP_WAIT"
	case JournalOpSignal:
		return "JOURNAL_OP_SIGNAL"
	default:
		return fmt.Sprintf("JOURNAL_OP_UNKNOWN(%d)", op)
	}
}
