package vm

import (
	"context"
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type IModule interface {
	Register(ctx context.Context) error
	Test(ctx context.Context) error
}

type Module struct {
	psi.NodeBase

	Name   string `json:"name,omitempty"`
	Source string `json:"source"`

	cached   *CachedModule
	instance *ModuleInstance
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

func (m *Module) Get(ctx context.Context) (*ModuleInstance, error) {
	if m.instance == nil {
		tx := coreapi.GetTransaction(ctx)
		vmctx := inject.Inject[*Context](tx.Graph().Services())

		lm, err := vmctx.Load(ctx, m)

		if err != nil {
			return nil, err
		}

		m.instance = lm
	}

	return m.instance, nil
}

func (m *Module) Register(ctx context.Context) (any, error) {
	lm, err := m.Get(ctx)

	if err != nil {
		return err, nil
	}

	r, err := lm.register(ctx, m)

	if err != nil {
		return err, nil
	}

	return r, nil
}

func (m *Module) Test(ctx context.Context) (any, error) {
	lm, err := m.Get(ctx)

	if err != nil {
		return err, nil
	}

	r, err := lm.Invoke("test", typesystem.ValueOf(&TestContext{}))

	if err != nil {
		return err, nil
	}

	return r, nil
}

type TestContext struct {
}

func (a *TestContext) Log(str string) {
	fmt.Printf("%s\n", str)
}
