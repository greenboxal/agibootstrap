package guifx

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Editor interface {
	Project() project.Project
	ElementPath() psi.Path
	Element() psi.Node
	Root() fyne.CanvasObject
}

type EditorFactory func(p project.Project, elementPath psi.Path, element psi.Node) Editor

type PsiNodeDescription struct {
	Name        string
	Description string
	Icon        fyne.Resource
}

func GetPsiNodeDescription(v psi.Node) PsiNodeDescription {
	switch v := v.(type) {
	case thoughtdb.Branch:
		return PsiNodeDescription{
			Name:        v.UUID(),
			Description: "Log Branch",
			Icon:        theme.AccountIcon(),
		}

	case *vfs.Directory:
		return PsiNodeDescription{
			Name:        v.PsiNodeName(),
			Description: v.String(),
			Icon:        theme.FolderIcon(),
		}

	case *vfs.FileNode:
		return PsiNodeDescription{
			Name:        v.PsiNodeName(),
			Description: v.String(),
			Icon:        theme.FileIcon(),
		}

	case psi.SourceFile:
		return PsiNodeDescription{
			Name:        fmt.Sprintf("Source File (%s)", v.Language().Name()),
			Description: v.String(),
			Icon:        theme.FileTextIcon(),
		}

	case psi.NamedNode:
		return PsiNodeDescription{
			Name:        v.PsiNodeName(),
			Description: v.String(),
			Icon:        theme.InfoIcon(),
		}
	}

	components := v.CanonicalPath().Components()
	baseName := components[len(components)-1]

	return PsiNodeDescription{
		Name:        baseName.String(),
		Description: v.String(),
		Icon:        theme.QuestionIcon(),
	}
}
