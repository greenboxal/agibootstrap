package visor

import (
	"sync"
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

	mu        sync.RWMutex
	pathCache map[string]*psiTreeNodeState

	SelectedItem   binding.Untyped
	OnNodeSelected func(n psi.Node)
}

func (ptw *PsiTreeWidget) Node(id widget.TreeNodeID) psi.Node {
	n, err := ptw.resolveCached(id)

	if err != nil {
		return nil
	}

	return n
}

func (ptw *PsiTreeWidget) getNodeState(id widget.TreeNodeID, create bool) *psiTreeNodeState {
	if id == "" {
		return nil
	}

	if state := ptw.pathCache[id]; state != nil {
		return state
	}

	ptw.mu.Lock()
	defer ptw.mu.Unlock()

	state := ptw.pathCache[id]

	if state == nil && create {
		p, err := psi.ParsePath(id)

		if err != nil {
			return nil
		}

		state = &psiTreeNodeState{
			tree: ptw,
			path: p,
		}

		ptw.pathCache[id] = state
	}

	return state
}

func (ptw *PsiTreeWidget) resolveCached(id string) (psi.Node, error) {
	state := ptw.getNodeState(id, true)

	if state.node == nil {
		state.loadNode()
	}

	return state.node, nil
}

func NewPsiTreeWidget(root psi.Node) *PsiTreeWidget {
	ptw := &PsiTreeWidget{
		root:      root,
		pathCache: map[string]*psiTreeNodeState{},

		SelectedItem: binding.NewUntyped(),
	}

	ptw.Tree = &widget.Tree{
		ChildUIDs: func(id widget.TreeNodeID) []widget.TreeNodeID {
			existing := ptw.getNodeState(id, true)

			if !existing.hasChildrenCached {
				existing.loadChildren()
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
			return NewPsiTreeItem().Container
		},

		UpdateNode: func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			n, err := ptw.resolveCached(id)

			if err != nil {
				return
			}

			info := GetPsiNodeDescription(n)

			labelContainer := o.(*fyne.Container)
			labelContainer.Objects[0].(*widget.Icon).SetResource(info.Icon)
			labelContainer.Objects[1].(*widget.Label).SetText(info.Name)
		},
	}

	ptw.Tree.OnBranchOpened = func(id widget.TreeNodeID) {
		entry := ptw.pathCache[id]

		if entry != nil {
			entry.invalidateChildren()

			ptw.Tree.Refresh()
		}
	}

	ptw.Tree.OnBranchClosed = func(id widget.TreeNodeID) {
		entry := ptw.pathCache[id]

		if entry != nil {
			entry.invalidateChildren()

			ptw.Tree.Refresh()
		}
	}

	ptw.Tree.OnSelected = func(id widget.TreeNodeID) {
		entry := ptw.pathCache[id]

		if entry != nil {
			entry.invalidateChildren()

			ptw.Tree.Refresh()
		}

		if ptw.OnNodeSelected != nil {
			ptw.OnNodeSelected(entry.node)
		}

		if err := ptw.SelectedItem.Set(entry.node); err != nil {
			panic(err)
		}
	}

	ptw.Tree.Root = ptw.root.CanonicalPath().String()

	ptw.Tree.ExtendBaseWidget(ptw.Tree)

	return ptw
}

type psiTreeNodeState struct {
	mu sync.RWMutex

	tree *PsiTreeWidget
	path psi.Path

	lastUpdate time.Time
	lastError  error

	node          psi.Node
	isNodeLoading bool

	children          []widget.TreeNodeID
	hasChildrenCached bool
	isChildrenLoading bool
}

func (s *psiTreeNodeState) loadNode() {
	if s.isNodeLoading {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isNodeLoading {
		return
	}

	s.isNodeLoading = true
	defer func() {
		s.isNodeLoading = false
	}()

	n, err := psi.ResolvePath(s.tree.root, s.path)

	if err != nil {
		return
	}

	if s.node != nil {
		s.node.RemoveInvalidationListener(s)
	}

	s.node = n

	if s.node != nil {
		s.node.AddInvalidationListener(s)
	}

	s.hasChildrenCached = false
	s.children = nil
}

func (s *psiTreeNodeState) loadChildren() bool {
	if s.isChildrenLoading {
		return false
	}

	if s.node == nil {
		s.loadNode()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isChildrenLoading {
		return false
	}

	s.isChildrenLoading = true
	defer func() {
		s.isChildrenLoading = false
	}()

	if s.node == nil {
		return false
	}

	ids := lo.Map(s.node.Children(), func(child psi.Node, _ int) widget.TreeNodeID {
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

	s.children = ids
	s.hasChildrenCached = true

	return true
}

func (s *psiTreeNodeState) invalidateChildren() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.hasChildrenCached = false
}

func (s *psiTreeNodeState) OnInvalidated(n psi.Node) {
	go func() {
		defer s.tree.Refresh()
		defer s.invalidateChildren()

		s.mu.Lock()
		defer s.mu.Unlock()

		s.node = n
	}()
}

func (s *psiTreeNodeState) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.node != nil {
		s.node.RemoveInvalidationListener(s)
	}

	s.node = nil
	s.children = nil
	s.hasChildrenCached = false
}

type PsiTreeItem struct {
	Container *fyne.Container

	Icon  *widget.Icon
	Label *widget.Label
}

func NewPsiTreeItem() *PsiTreeItem {
	ti := &PsiTreeItem{}

	ti.Icon = widget.NewIcon(theme.DocumentIcon())
	ti.Label = widget.NewLabel("Template Object")

	ti.Container = container.NewHBox(ti.Icon, ti.Label)

	return ti
}
