package visor

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/visor/guifx"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

func FactoryForNode(element psi.Node) guifx.EditorFactory {
	switch element.(type) {
	case thoughtdb.Branch:
		return NewThoughtLogEditor
	case project.SourceFile:
		return NewSourceFileEditor
	}

	return nil
}
