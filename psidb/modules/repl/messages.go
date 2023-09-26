package repl

import (
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type EvalCommandMessage struct {
	coreapi.SessionMessageBase
}

func init() {
	typesystem.GetType[EvalCommandMessage]()
}
