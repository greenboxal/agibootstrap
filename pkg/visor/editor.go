package visor

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtstream"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Editor interface {
	Project() project.Project
	ElementPath() psi.Path
	Element() psi.Node
	Root() fyne.CanvasObject
}

type EditorFactory func(p project.Project, elementPath psi.Path, element psi.Node) Editor

type PsiNodeType struct {
	Name string
	Icon fyne.Resource
}

func GetPsiNodeType(v psi.Node) PsiNodeType {
	switch v.(type) {
	case *thoughtstream.ThoughtLog:
		return PsiNodeType{
			Name: "Thought Log",
			Icon: theme.AccountIcon(),
		}

	case psi.SourceFile:
		return PsiNodeType{
			Name: "Source File",
			Icon: theme.DocumentIcon(),
		}
	}

	return PsiNodeType{
		Name: fmt.Sprintf("%T", v),
	}
}

func FactoryForNode(element psi.Node) EditorFactory {
	switch element.(type) {
	case *thoughtstream.ThoughtLog:
		return NewThoughtLogEditor
	}

	return nil
}
