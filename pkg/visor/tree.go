package visor

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	collectionsfx2 "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type PsiTreeWidget struct {
	*widget.Tree

	resolutionRoot psi.Node

	mu        sync.RWMutex
	pathCache map[string]*psiTreeNodeState
	refresher *debouncer

	OnNodeSelected func(n psi.Node)
}

func NewPsiTreeWidget(resolutionRoot psi.Node) *PsiTreeWidget {
	ptw := &PsiTreeWidget{
		resolutionRoot: resolutionRoot,
		pathCache:      map[string]*psiTreeNodeState{},
	}

	ptw.Tree = &widget.Tree{
		ChildUIDs: func(id widget.TreeNodeID) []widget.TreeNodeID {
			existing := ptw.getNodeState(id, true)

			if existing.node == nil {
				existing.loadNode()
			}

			return existing.childrenIds.Slice()
		},

		IsBranch: func(id widget.TreeNodeID) bool {
			return true
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

	ptw.Tree.OnSelected = func(id widget.TreeNodeID) {
		entry := ptw.pathCache[id]

		if ptw.OnNodeSelected != nil {
			ptw.OnNodeSelected(entry.node)
		}
	}

	ptw.Tree.ExtendBaseWidget(ptw.Tree)

	ptw.refresher = newDebouncer(500*time.Millisecond, func() {
		ptw.Tree.Refresh()
	})

	ptw.SetRootItem(resolutionRoot.CanonicalPath())

	return ptw
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

		state = newPsiTreeNodeState(ptw, p)

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

func (ptw *PsiTreeWidget) SetRootItem(path psi.Path) {
	state := ptw.getNodeState(path.String(), true)

	if state.node == nil {
		state.loadNode()
	}

	ptw.Tree.Root = path.String()
}

func (ptw *PsiTreeWidget) refreshTree() {
	ptw.refresher.Request()
}

type psiTreeNodeState struct {
	mu sync.RWMutex

	tree *PsiTreeWidget
	path psi.Path

	lastUpdate time.Time
	lastError  error

	node          psi.Node
	isNodeLoading bool

	childrenIds collectionsfx2.MutableSlice[widget.TreeNodeID]
}

func newPsiTreeNodeState(ptw *PsiTreeWidget, p psi.Path) *psiTreeNodeState {
	st := &psiTreeNodeState{
		tree: ptw,
		path: p,
	}

	st.childrenIds.AddListener(obsfx.OnInvalidatedFunc(func(o obsfx.Observable) {
		ptw.refreshTree()
	}))

	return st
}

type debouncer struct {
	mu      sync.Mutex
	f       func()
	size    time.Duration
	last    time.Time
	waiting bool
}

func newDebouncer(windowSize time.Duration, f func()) *debouncer {
	return &debouncer{size: windowSize, f: f}
}

func (d *debouncer) Request() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.waiting {
		return
	}

	d.waiting = true
	defer func() {
		d.waiting = false
	}()

	go func() {
		time.Sleep(d.size)

		d.mu.Lock()
		defer d.mu.Unlock()

		if d.waiting {
			d.f()
		}
	}()
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

	n, err := psi.ResolvePath(s.tree.resolutionRoot, s.path)

	if err != nil {
		return
	}

	if s.node != n {
		s.childrenIds.Clear()

		if s.node != nil {
			s.childrenIds.Unbind()
		}

		s.node = n

		if s.node != nil {
			collectionsfx2.BindList(&s.childrenIds, s.node.ChildrenList(), func(v psi.Node) widget.TreeNodeID {
				return v.CanonicalPath().String()
			})
		}
	}
}

func (s *psiTreeNodeState) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.childrenIds.Unbind()
	s.node = nil
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
