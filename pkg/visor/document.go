package visor

import (
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
	p         project.Project
	area      *DocumentArea
	documents map[string]*Document
}

func (m *DocumentManager) AddDocument(doc *Document) {
	m.documents[doc.ID] = doc
	m.area.Append(doc.tabItem)
}

func (m *DocumentManager) CloseDocument(d *Document) {
	delete(m.documents, d.ID)

	m.area.Remove(d.tabItem)
}

func (m *DocumentManager) OpenDocument(path psi.Path, node psi.Node) {
	key := path.String()
	existing := m.documents[key]

	if existing != nil {
		m.area.Select(existing.tabItem)
		return
	}

	if node == nil {
		n, err := psi.ResolvePath(m.p, path)

		if err != nil {
			panic(err)
		}

		node = n
	}

	factory := FactoryForNode(node)

	if factory == nil {
		return
	}

	editor := factory(m.p, path, node)

	m.AddDocument(NewDocument(path.String(), path.String(), editor.Root()))
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

	m.area = da

	return da
}
