package thoughtstream

import (
	"github.com/hashicorp/go-multierror"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Manager struct {
	psi.NodeBase

	basePath string
}

func NewManager(basePath string) *Manager {
	lm := &Manager{
		basePath: basePath,
	}

	lm.NodeBase.Init(lm, "")

	return lm
}

func (m *Manager) PsiNodeName() string { return "ThoughtStreamManager" }

func (m *Manager) GetOrCreateStream(name string) (log *ThoughtLog, err error) {
	key := psi.TypedEdgeKey[*ThoughtLog]{
		Kind: psi.TypedEdgeKind[*ThoughtLog](psi.EdgeKindChild),
		Name: name,
	}

	log = psi.ResolveEdge(m.PsiNode(), key)

	if log == nil {
		log, err = NewThoughtLog(name, m.basePath)

		if err != nil {
			return nil, err
		}

		log.SetParent(m.PsiNode())
	}

	return log, nil
}

func (m *Manager) Close() error {
	var err error

	for _, m := range m.PsiNode().Children() {
		m, ok := m.(*ThoughtLog)

		if !ok {
			continue
		}

		if e := m.Close(); e != nil {
			err = multierror.Append(err, e)
		}
	}

	return nil
}
