package visor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/samber/lo"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type PsiTreeWidget struct {
	*widget.Tree

	root psi.Node

	pathCache  map[string]psi.Node
	childCache map[widget.TreeNodeID][]widget.TreeNodeID

	SelectedItem binding.Untyped
}

func (w *PsiTreeWidget) Node(id widget.TreeNodeID) psi.Node {
	n, err := w.resolveCached(id)

	if err != nil {
		return nil
	}

	return n
}

func (w *PsiTreeWidget) resolveCached(id string) (psi.Node, error) {
	if id == "" {
		panic("empty id")
	}

	if n := w.pathCache[id]; n != nil {
		return n, nil
	}

	p := psi.MustParsePath(id)

	n, err := psi.ResolvePath(w.root, p)

	if err != nil {
		return nil, err
	}

	w.pathCache[id] = n

	return n, nil
}

func (w *PsiTreeWidget) refreshNode(id widget.TreeNodeID) {
	delete(w.pathCache, id)
	delete(w.childCache, id)

	w.Tree.Refresh()
}

func NewPsiTreeWidget(root psi.Node) *PsiTreeWidget {
	ptw := &PsiTreeWidget{
		root:       root,
		pathCache:  map[string]psi.Node{},
		childCache: map[widget.TreeNodeID][]widget.TreeNodeID{},

		SelectedItem: binding.NewUntyped(),
	}

	ptw.Tree = &widget.Tree{
		ChildUIDs: func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				return []widget.TreeNodeID{"/"}
			}

			if existing, ok := ptw.childCache[id]; ok {
				return existing
			}

			resolved, err := ptw.resolveCached(id)

			if err != nil {
				return nil
			}

			ids := lo.Map(resolved.Children(), func(child psi.Node, _ int) widget.TreeNodeID {
				p := child.CanonicalPath()

				if p.IsEmpty() {
					panic("empty path")
				}

				ps := p.String()

				if ps == "" {
					panic("empty path")
				}

				return ps
			})

			ptw.childCache[id] = ids

			return ids
		},

		IsBranch: func(id widget.TreeNodeID) bool {
			if id == "" {
				return true
			}

			n, err := ptw.resolveCached(id)

			if err != nil {
				return false
			}

			return n.IsContainer()
		},

		CreateNode: func(branch bool) fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("Template Object"))
		},

		UpdateNode: func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			if id == "" {
				return
			}

			n, err := ptw.resolveCached(id)

			if err != nil {
				return
			}

			text := ""

			if named, ok := n.(psi.NamedNode); ok {
				text = named.PsiNodeName()
			} else {
				text = n.String()
			}

			info := GetPsiNodeType(n)

			o.(*fyne.Container).Objects[0].(*widget.Icon).SetResource(info.Icon)
			o.(*fyne.Container).Objects[1].(*widget.Label).SetText(text)
		},
	}

	ptw.Tree.OnBranchOpened = func(id widget.TreeNodeID) {
		ptw.refreshNode(id)
	}

	ptw.Tree.OnSelected = func(id widget.TreeNodeID) {
		ptw.refreshNode(id)

		if err := ptw.SelectedItem.Set(ptw.Node(id)); err != nil {
			panic(err)
		}
	}

	ptw.Tree.ExtendBaseWidget(ptw.Tree)

	return ptw
}
