package coreapi

import (
	"context"
	"time"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type SessionConfig struct {
	SessionID       string `json:"session_id"`
	ParentSessionID string `json:"parent_session_id,omitempty"`

	Root          psi.Path            `json:"root"`
	Journal       JournalConfig       `json:"journal"`
	Checkpoint    CheckpointConfig    `json:"checkpoint"`
	MetadataStore MetadataStoreConfig `json:"metadata_store"`
	MountPoints   []MountDefinition   `json:"mount_points"`

	Persistent       bool          `json:"persistent"`
	KeepAliveTimeout time.Duration `json:"keep_alive_timeout"`
	Deadline         time.Time     `json:"deadline"`
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

	if mixin.MetadataStore != nil {
		c.MetadataStore = mixin.MetadataStore
	}

	c.MountPoints = append(c.MountPoints, mixin.MountPoints...)

	if mixin.KeepAliveTimeout > 0 {
		c.KeepAliveTimeout = mixin.KeepAliveTimeout
	}

	if !mixin.Deadline.IsZero() {
		c.Deadline = mixin.Deadline
	}

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
