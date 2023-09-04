package journal

import (
	`context`

	coreapi `github.com/greenboxal/agibootstrap/psidb/core/api`
)

type FileJournalConfig struct {
	Path string `json:"path"`
}

func (f FileJournalConfig) CreateJournal(ctx context.Context) (coreapi.Journal, error) {
	return OpenJournal(f.Path)
}
