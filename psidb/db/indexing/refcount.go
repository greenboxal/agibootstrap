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

func (rci *referenceCountingIndex) Close() error {
	rci.refMutex.Lock()
	defer rci.refMutex.Unlock()

	if rci.refs.Load() > 0 {
		rci.faissIndex.m.logger.Warn("Closing index with active references")
	}

	return rci.faissIndex.Close()
}

type indexReference struct {
	*referenceCountingIndex

	closed bool
}

func (ir *indexReference) Close() error {
	ir.refMutex.Lock()
	defer ir.refMutex.Unlock()

	if ir.closed {
		return nil
	}

	if err := ir.faissIndex.Save(); err != nil {
		return err
	}

	if ir.refs.Add(-1) > 0 {
		ir.faissIndex.m.notifyIndexIdle(ir.faissIndex.id)
	}

	ir.closed = true

	return nil
}
