package coreapi

import (
	`context`

	`github.com/greenboxal/agibootstrap/pkg/psi`
)

type SessionConfig struct {
	SessionID string `json:"session_id"`

	Root          psi.Path            `json:"root"`
	Journal       JournalConfig       `json:"journal"`
	Checkpoint    CheckpointConfig    `json:"checkpoint"`
	LinkedStore   LinkedStoreConfig   `json:"blob_store"`
	MetadataStore MetadataStoreConfig `json:"metadata_store"`
	MountPoints   []MountDefinition   `json:"mount_points"`
}

func (c SessionConfig) Extend(mixin SessionConfig) SessionConfig {
	c.SessionID = mixin.SessionID

	if mixin.Root.IsEmpty() || mixin.Root.IsRelative() {
		c.Root = c.Root.Join(mixin.Root)
	} else {
		c.Root = mixin.Root
	}

	if mixin.Journal != nil {
		c.Journal = mixin.Journal
	}

	if mixin.Checkpoint != nil {
		c.Checkpoint = mixin.Checkpoint
	}

	if mixin.LinkedStore != nil {
		c.LinkedStore = mixin.LinkedStore
	}

	if mixin.MetadataStore != nil {
		c.MetadataStore = mixin.MetadataStore
	}

	c.MountPoints = append(c.MountPoints, mixin.MountPoints...)

	return c
}

type MountTab map[string]MountDefinition

func NewMountTab(definitions ...MountDefinition) MountTab {
	m := map[string]MountDefinition{}

	for _, def := range definitions {
		m[def.Path.String()] = def
	}

	return m
}

type MountTarget interface {
	Mount(ctx context.Context, md MountDefinition) (any, error)
}
