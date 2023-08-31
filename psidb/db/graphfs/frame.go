package graphfs

import "github.com/greenboxal/agibootstrap/psidb/core/api"

type Frame struct {
	Log []coreapi.JournalEntry `json:"log"`
}
