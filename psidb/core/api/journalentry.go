package coreapi

import (
	"fmt"

	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

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
