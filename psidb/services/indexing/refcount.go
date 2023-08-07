package indexing

import (
	"sync"
	"sync/atomic"
)

type referenceCountingIndex struct {
	*faissIndex

	refMutex sync.Mutex
	refs     atomic.Int64
}

func (rci *referenceCountingIndex) ref() (BasicIndex, error) {
	rci.refMutex.Lock()
	defer rci.refMutex.Unlock()

	rci.refs.Add(1)

	return &indexReference{
		referenceCountingIndex: rci,
	}, nil
}

func (rci *referenceCountingIndex) Close(force bool) error {
	rci.refMutex.Lock()
	defer rci.refMutex.Unlock()

	if rci.refs.Load() > 0 {
		if force {
			rci.faissIndex.m.logger.Warn("Closing index with active references")
		} else {
			return nil
		}
	}

	return rci.faissIndex.Close()
}

type indexReference struct {
	*referenceCountingIndex

	closed bool
}

func (ir *indexReference) Close() error {
	ir.refMutex.Lock()

	if ir.closed {
		ir.refMutex.Unlock()
		return nil
	}

	ir.closed = true
	refCount := ir.refs.Add(-1)
	ir.refMutex.Unlock()

	if refCount <= 0 {
		ir.faissIndex.m.notifyIndexIdle(ir.faissIndex)
	}

	return nil
}
