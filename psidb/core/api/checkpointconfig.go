package coreapi

import (
	`context`
)

type CheckpointConfig interface {
	CreateCheckpoint(ctx context.Context) (Checkpoint, error)
}

type FileCheckpointConfig struct {
	Path string `json:"path"`
}

func (f FileCheckpointConfig) CreateCheckpoint(ctx context.Context) (Checkpoint, error) {
	return OpenFileCheckpoint(f.Path)
}

type InMemoryCheckpointConfig struct{}

func (i InMemoryCheckpointConfig) CreateCheckpoint(ctx context.Context) (Checkpoint, error) {
	return NewInMemoryCheckpoint(), nil
}
