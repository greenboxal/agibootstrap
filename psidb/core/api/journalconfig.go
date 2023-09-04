package coreapi

import (
	`context`

	`github.com/greenboxal/agibootstrap/pkg/psi`
)

type JournalConfig interface {
	CreateJournal(ctx context.Context) (Journal, error)
}

type MountDefinition struct {
	Name   string      `json:"name"`
	Path   psi.Path    `json:"path"`
	Target MountTarget `json:"target,omitempty"`
}
