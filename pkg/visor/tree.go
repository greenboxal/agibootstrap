package visor

import (
	"time"

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

	pathCache map[string]*psiTreeNodeState

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

	existing := w.pathCache[id]

	if existing == nil {
		existing = &psiTreeNodeState{}

		w.pathCache[id] = existing
	}

	if existing.node == nil {
		p := psi.MustParsePath(id)

		n, err := psi.ResolvePath(w.root, p)

		if err != nil {
			return nil, err
		}

		existing.node = n
	}

	return existing.node, nil
}

func (w *PsiTreeWidget) refreshNode(id widget.TreeNodeID) {
	delete(w.pathCache, id)

	w.Tree.Refresh()
}

type psiTreeNodeState struct {
	ts       time.Time
	node     psi.Node
	children []widget.TreeNodeID
	valid    bool
}

func NewPsiTreeWidget(root psi.Node) *PsiTreeWidget {
	ptw := &PsiTreeWidget{
		root:      root,
		pathCache: map[string]*psiTreeNodeState{},

		SelectedItem: binding.NewUntyped(),
	}

	invalidationListener := psi.InvalidationListenerFunc(func(n psi.Node) {
		entry := ptw.pathCache[n.CanonicalPath().String()]

		if entry != nil {
			entry.valid = false
		}

		ptw.Tree.Refresh()
	})

	ptw.Tree = &widget.Tree{
		ChildUIDs: func(id widget.TreeNodeID) []widget.TreeNodeID {
			existing := ptw.pathCache[id]

			if existing == nil {
				existing = &psiTreeNodeState{}

				ptw.pathCache[id] = existing
			}

			if existing.node == nil {
				resolved, err := ptw.resolveCached(id)

				if err != nil {
					return nil
				}

				existing.node = resolved
			}

			if !existing.valid {
				go func() {
					ids := lo.Map(existing.node.Children(), func(child psi.Node, _ int) widget.TreeNodeID {
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

					existing.children = ids
					existing.valid = true

					ptw.Tree.Refresh()
				}()
			}

			return existing.children
		},

		IsBranch: func(id widget.TreeNodeID) bool {
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
		entry := ptw.pathCache[id]

		if entry != nil && entry.node != nil {
			entry.node.AddInvalidationListener(invalidationListener)
		}

		go ptw.Tree.Refresh()
	}

	ptw.Tree.OnBranchClosed = func(id widget.TreeNodeID) {
		entry := ptw.pathCache[id]

		if entry != nil {
			if entry.node != nil {
				entry.node.RemoveInvalidationListener(invalidationListener)
			}

			entry.valid = false
		}
	}

	ptw.Tree.OnSelected = func(id widget.TreeNodeID) {
		entry := ptw.pathCache[id]

		if entry != nil {
			entry.valid = false
		}

		if err := ptw.SelectedItem.Set(ptw.Node(id)); err != nil {
			panic(err)
		}
	}

	ptw.Tree.Root = ptw.root.CanonicalPath().String()

	ptw.Tree.ExtendBaseWidget(ptw.Tree)

	return ptw
}
