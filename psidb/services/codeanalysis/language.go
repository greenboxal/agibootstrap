package codeanalysis

import (
	"context"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type Language interface {
	Parse(ctx context.Context, src *SourceFile) (CompilationUnit, error)
}

type CompilationUnit interface {
	GetRoot() psi.Node
	GetScope() *Scope
}
