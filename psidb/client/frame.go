package client

import (
	"time"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Frame struct {
	Log []coreapi.JournalEntry `json:"log"`
}

type FrameBuilder struct {
	log []coreapi.JournalEntry

	ridCounter uint64
}

func NewFrameBuilder() *FrameBuilder {
	return &FrameBuilder{}
}

func (fb *FrameBuilder) IsEmpty() bool {
	return len(fb.log) == 0
}

func (fb *FrameBuilder) Add(entry coreapi.JournalEntry) *FrameBuilder {
	if entry.Rid == 0 {
		fb.ridCounter++
		entry.Rid = fb.ridCounter
	} else if entry.Rid > fb.ridCounter {
		fb.ridCounter = entry.Rid
	}

	if entry.Ts == 0 {
		entry.Ts = time.Now().UnixNano()
	}

	fb.log = append(fb.log, entry)

	return fb
}

func (fb *FrameBuilder) Reset() *FrameBuilder {
	fb.log = nil
	fb.ridCounter = 0

	return fb
}

func (fb *FrameBuilder) Build() *Frame {
	return &Frame{
		Log: fb.log,
	}
}
