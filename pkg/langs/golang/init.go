package golang

import (
	"github.com/greenboxal/agibootstrap/pkg/codex"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

const LanguageID psi.LanguageID = "go"

func init() {
	codex.RegisterLanguage(LanguageID, NewLanguage)
}
