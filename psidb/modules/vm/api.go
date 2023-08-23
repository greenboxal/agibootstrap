package vm

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type IModule interface {
	Register(ctx context.Context) error
}

type Module struct {
	psi.NodeBase

	Name   string `json:"name,omitempty"`
	Source string `json:"source"`

	cached *CachedModule
	lm     *LiveModule
}

var ModuleInterface = psi.DefineNodeInterface[IModule]()
var ModuleType = psi.DefineNodeType[*Module](
	psi.WithInterfaceFromNode(ModuleInterface),
)

func NewModule(name string, source string) *Module {
	m := &Module{
		Name:   name,
		Source: source,
	}

	m.Init(m, psi.WithNodeType(ModuleType))

	return m
}

func (m *Module) PsiNodeName() string { return m.Name }

func (m *Module) Get(ctx context.Context) (*LiveModule, error) {
	if m.lm == nil {
		tx := coreapi.GetTransaction(ctx)
		vmctx := inject.Inject[*Context](tx.Graph().Services())

		lm, err := vmctx.Load(ctx, m)

		if err != nil {
			return nil, err
		}

		m.lm = lm
	}

	return m.lm, nil
}

func (m *Module) Register(ctx context.Context) (any, error) {
	lm, err := m.Get(ctx)

	if err != nil {
		return err, nil
	}

	r, err := lm.Invoke("register")

	if err != nil {
		return err, nil
	}

	return r.Value().Interface(), nil
}
