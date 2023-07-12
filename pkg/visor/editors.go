package visor

import (
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/langs"
	"github.com/greenboxal/agibootstrap/pkg/visor/guifx"
)

func FactoryForNode(element psi.Node) guifx.EditorFactory {
	switch element.(type) {
	case thoughtdb.Branch:
		return NewThoughtLogEditor
	case langs.SourceFile:
		return NewSourceFileEditor
	}

	return nil
}
