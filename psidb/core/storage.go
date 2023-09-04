package core

import (
	"context"
	"os"
	"path"

	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/psidb/core/api"
	graphfs `github.com/greenboxal/agibootstrap/psidb/db/journal`
)

func NewJournal(
	cfg *coreapi.Config,
	lc fx.Lifecycle,
) (*graphfs.Journal, error) {
	walPath := path.Join(cfg.DataDir, "wal")

	if err := os.MkdirAll(walPath, 0755); err != nil {
		return nil, err
	}

	journal, err := graphfs.OpenJournal(walPath)

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return journal.Close()
		},
	})

	return journal, nil
}

func NewCheckpoint(
	cfg *coreapi.Config,
	lc fx.Lifecycle,
) (coreapi.Checkpoint, error) {
	ckptPath := path.Join(cfg.DataDir, "ckpt.bin")

	if err := os.MkdirAll(path.Dir(ckptPath), 0755); err != nil {
		return nil, err
	}

	checkpoint, err := coreapi.OpenFileCheckpoint(ckptPath)

	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return checkpoint.Close()
		},
	})

	return checkpoint, nil
}
