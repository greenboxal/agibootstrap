package visor

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Document struct {
	ID      string
	Name    string
	Content fyne.CanvasObject

	manager *DocumentManager
	tabItem *container.TabItem
	isOpen  bool
}

func (d *Document) Close() {
	d.manager.CloseDocument(d)
}

func NewDocument(id, name string, content fyne.CanvasObject) *Document {
	return &Document{
		ID:      id,
		Name:    name,
		Content: content,

		tabItem: container.NewTabItem(name, content),
	}
}

type DocumentManager struct {
	mu        sync.Mutex
	p         project.Project
	area      *DocumentArea
	documents map[string]*Document
}

func (m *DocumentManager) AddDocument(doc *Document) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.documents[doc.ID] != nil {
		return
	}

	m.documents[doc.ID] = doc

	if !doc.isOpen {
		m.area.Append(doc.tabItem)

		m.area.Refresh()

		doc.isOpen = true
	}
}

func (m *DocumentManager) CloseDocument(d *Document) {
	m.mu.Lock()
	defer m.mu.Unlock()

	d.isOpen = false

	m.area.Remove(d.tabItem)
	m.area.Refresh()
}

func (m *DocumentManager) OpenDocument(path psi.Path, node psi.Node) {
	key := path.String()

	if node == nil {
		n, err := psi.ResolvePath(m.p, path)

		if err != nil {
			panic(err)
		}

		node = n
	}

	existing := m.documents[key]

	if existing == nil {
		factory := FactoryForNode(node)

		if factory == nil {
			return
		}

		editor := factory(m.p, path, node)

		existing = NewDocument(path.String(), path.String(), editor.Root())
	}

	if !existing.isOpen {
		m.AddDocument(existing)
	}

	m.area.Select(existing.tabItem)
}

func NewDocumentManager(p project.Project) *DocumentManager {
	dm := &DocumentManager{
		documents: map[string]*Document{},
	}

	dm.area = NewDocumentArea(dm)

	return dm
}

type DocumentArea struct {
	*container.DocTabs

	manager *DocumentManager
}

func NewDocumentArea(m *DocumentManager) *DocumentArea {
	if m.area != nil {
		return m.area
	}

	da := &DocumentArea{
		DocTabs: container.NewDocTabs(),

		manager: m,
	}

	da.DocTabs.OnClosed = func(tab *container.TabItem) {
		m.mu.Lock()
		defer m.mu.Unlock()

		for _, doc := range m.documents {
			if doc.tabItem == tab {
				da.manager.CloseDocument(doc)
				return
			}
		}
	}

	m.area = da

	return da
}
