package guifx

import (
	"context"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bep/debounce"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	collectionsfx "github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

var logger = logging.GetLogger("psi-tree-widget")

type PsiTreeWidget struct {
	*widget.Tree

	resolutionRoot *graphstore.IndexedGraph

	mu        sync.RWMutex
	pathCache map[string]*psiTreeNodeState

	updateDebouncer func(f func())

	OnNodeSelected func(n psi.Node)
}

func NewPsiTreeWidget(resolutionRoot *graphstore.IndexedGraph) *PsiTreeWidget {
	ptw := &PsiTreeWidget{
		resolutionRoot: resolutionRoot,

		pathCache: map[string]*psiTreeNodeState{},

		updateDebouncer: debounce.New(100 * time.Millisecond),
	}

	ptw.Tree = &widget.Tree{
		ChildUIDs: func(id widget.TreeNodeID) []widget.TreeNodeID {
			existing := ptw.getNodeState(id, true)

			if existing == nil {
				return nil
			}

			if existing.Node() == nil {
				existing.loadNode(context.Background())
			}

			return existing.childrenIdsCached
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

	ptw.Tree.OnBranchOpened = func(id widget.TreeNodeID) {
		existing := ptw.getNodeState(id, true)

		existing.loadChildren()
	}

	ptw.Tree.OnBranchClosed = func(id widget.TreeNodeID) {
		existing := ptw.getNodeState(id, false)

		if existing == nil {
			return
		}

		existing.unloadChildren()
	}

	ptw.Tree.OnSelected = func(id widget.TreeNodeID) {
		entry := ptw.pathCache[id]

		if ptw.OnNodeSelected != nil {
			ptw.OnNodeSelected(entry.Node())
		}
	}

	ptw.Tree.ExtendBaseWidget(ptw.Tree)

	return ptw
}

func (ptw *PsiTreeWidget) SetRootItem(path psi.Path) {
	ptw.Tree.Root = path.String()

	entry := ptw.getNodeState(path.String(), true)
	entry.loadChildren()

	ptw.Refresh()
}

func (ptw *PsiTreeWidget) Node(id widget.TreeNodeID) psi.Node {
	n, err := ptw.resolveCached(id)

	if err != nil {
		return nil
	}

	return n
}

func (ptw *PsiTreeWidget) resolveCached(id string) (psi.Node, error) {
	state := ptw.getNodeState(id, true)

	if state.Node() == nil {
		state.loadNode(context.Background())
	}

	return state.Node(), nil
}

func (ptw *PsiTreeWidget) getNodeState(id widget.TreeNodeID, create bool) *psiTreeNodeState {
	if id == "" {
		return nil
	}

	ptw.mu.Lock()
	defer ptw.mu.Unlock()

	if state := ptw.pathCache[id]; state != nil {
		return state
	}

	if !create {
		return nil
	}

	p, err := psi.ParsePath(id)

	if err != nil {
		return nil
	}

	state := newPsiTreeNodeState(ptw, p)

	ptw.pathCache[id] = state

	return state
}

func (ptw *PsiTreeWidget) refreshTree() {
	ptw.updateDebouncer(func() {
		ptw.Tree.Refresh()
	})
}

type psiTreeNodeState struct {
	mu sync.RWMutex

	tree *PsiTreeWidget
	path psi.Path

	lastUpdate time.Time
	lastError  error

	isNodeLoading    bool
	isChildrenLoaded bool

	node        obsfx.SimpleProperty[psi.Node]
	childrenIds collectionsfx.MutableSlice[widget.TreeNodeID]

	childrenIdsCached []widget.TreeNodeID
}

func (s *psiTreeNodeState) Node() psi.Node { return s.node.Value() }

func newPsiTreeNodeState(ptw *PsiTreeWidget, p psi.Path) *psiTreeNodeState {
	st := &psiTreeNodeState{
		tree: ptw,
		path: p,
	}

	obsfx.ObserveChange(&st.node, func(old, new psi.Node) {
		wasLoaded := st.isChildrenLoaded

		st.unloadChildren()

		if new != nil && wasLoaded {
			st.loadChildren()
		}
	})

	st.childrenIds.AddListener(obsfx.OnInvalidatedFunc(func(o obsfx.Observable) {
		self := st.path.String()

		it := iterators.Filter(iterators.FromSlice(st.childrenIds.Slice()), func(t widget.TreeNodeID) bool {
			return t != self && strings.HasPrefix(t, self)
		})

		st.childrenIdsCached = iterators.ToSlice(it)

		ptw.refreshTree()
	}))

	return st
}

func (s *psiTreeNodeState) loadChildren() {
	if s.isChildrenLoaded {
		return
	}

	s.childrenIdsCached = nil
	s.childrenIds.Clear()
	s.childrenIds.Unbind()

	if s.Node() != nil {
		collectionsfx.BindList(&s.childrenIds, s.Node().ChildrenList(), func(v psi.Node) widget.TreeNodeID {
			return v.CanonicalPath().String()
		})
	}

	s.isChildrenLoaded = true
}

func (s *psiTreeNodeState) unloadChildren() {
	if !s.isChildrenLoaded {
		return
	}

	s.isChildrenLoaded = false
	s.childrenIds.Unbind()
	s.childrenIds.Clear()
	s.childrenIdsCached = nil
}

func (s *psiTreeNodeState) loadNode(ctx context.Context) {
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

	n, err := s.tree.resolutionRoot.ResolveNode(ctx, s.path)

	if err != nil {
		logger.Error(err)
		return
	}

	s.node.SetValue(n)
}

func (s *psiTreeNodeState) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.node.SetValue(nil)
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
