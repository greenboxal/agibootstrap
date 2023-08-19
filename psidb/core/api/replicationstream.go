package coreapi

import (
	"context"
	"io"
	"time"

	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/psidb/db/graphfs"
)

type ReplicationStreamProcessorFunc func(ctx context.Context, entry []*graphfs.JournalEntry) error

type ReplicationStreamProcessor struct {
	logger  *zap.SugaredLogger
	slot    graphfs.ReplicationSlot
	process ReplicationStreamProcessorFunc
	proc    goprocess.Process
	running bool
}

func NewReplicationStream(slot graphfs.ReplicationSlot, processFn ReplicationStreamProcessorFunc) *ReplicationStreamProcessor {
	rsp := &ReplicationStreamProcessor{
		slot:   slot,
		logger: logging.GetLogger("replication-stream:" + slot.Name()),
	}

	rsp.proc = goprocess.Go(rsp.run)
	rsp.process = processFn

	return rsp
}

func (s *ReplicationStreamProcessor) run(proc goprocess.Process) {
	ctx := goprocessctx.OnClosingContext(proc)

	buffer := make([]graphfs.ReplicationMessage, 16)

	s.running = true

	for s.running {
		n, err := s.slot.Read(ctx, buffer)

		if err == io.EOF || n == 0 {
			time.Sleep(100 * time.Millisecond)

			continue
		} else if err != nil {
			s.logger.Error(err)

			panic(err)
		}

		for i := 0; i < n; i++ {
			if err := s.process(ctx, buffer[i].Entries); err != nil {
				s.logger.Error(err)

				panic(err)
			}

			if err := s.Commit(ctx); err != nil {
				s.logger.Error(err)

				panic(err)
			}
		}
	}
}

func (s *ReplicationStreamProcessor) Commit(ctx context.Context) error {
	if err := s.slot.FlushPosition(ctx); err != nil {
		return err
	}

	return nil
}

func (s *ReplicationStreamProcessor) Close(ctx context.Context) error {
	s.running = false

	if err := s.proc.CloseAfterChildren(); err != nil {
		return err
	}

	return s.Commit(ctx)
}
