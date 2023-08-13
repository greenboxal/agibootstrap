package typing

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
)

type Manager struct {
	core     coreapi.Core
	migrator migrations.Migrator

	intrinsicRegistry psi.TypeRegistry
}

func NewManager(
	lc fx.Lifecycle,
	core coreapi.Core,
	migrator migrations.Migrator,
) *Manager {
	m := &Manager{
		core:              core,
		migrator:          migrator,
		intrinsicRegistry: psi.GlobalTypeRegistry(),
	}

	lc.Append(fx.Hook{
		OnStart: m.Start,
	})

	return m
}

func (m *Manager) Start(ctx context.Context) error {
	if err := m.migrator.Migrate(ctx, migrationSet); err != nil {
		return err
	}

	return m.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		for _, nt := range m.intrinsicRegistry.NodeTypes() {
			if _, err := m.registerNodeType(ctx, nt); err != nil {
				return err
			}
		}

		return nil
	})
}

func (m *Manager) registerNodeType(ctx context.Context, nt psi.NodeType) (*Type, error) {
	pkgComponents := strings.Split(nt.Name(), ".")
	pkgComponents = pkgComponents[:len(pkgComponents)-1]
	pkgName := strings.Join(pkgComponents, ".")

	pkg, err := m.lookupPackage(ctx, pkgName, true)

	if err != nil {
		return nil, err
	}

	tn := pkg.ResolveChild(ctx, psi.PathElement{Name: nt.Name()})

	if tn == nil {
		t := NewType(nt.Name())
		t.SetParent(pkg)

		if err := t.Update(ctx); err != nil {
			return nil, err
		}

		tn = t
	}

	return tn.(*Type), nil
}

func (m *Manager) lookupType(ctx context.Context, name string) (resolved *Type, err error) {
	pkgComponents := strings.Split(name, ".")
	pkgComponents = pkgComponents[:len(pkgComponents)-1]
	pkgName := strings.Join(pkgComponents, ".")

	pkg, err := m.lookupPackage(ctx, pkgName, true)

	if err != nil {
		return nil, err
	}

	tn := pkg.ResolveChild(ctx, psi.PathElement{Name: name})

	if tn == nil {
		return nil, nil
	}

	return tn.(*Type), nil
}

func (m *Manager) lookupPackage(ctx context.Context, name string, create bool) (resolved *Package, err error) {
	tx := coreapi.GetTransaction(ctx)

	pkgs := strings.Split(name, ".")

	elements := lo.Map(pkgs, func(pkg string, _ int) psi.PathElement {
		return psi.PathElement{Name: pkg}
	})

	p := RootPath.Join(psi.PathFromElements("", false, elements...))

	resolved, err = psi.Resolve[*Package](ctx, tx.Graph(), p)

	if err != nil && (!create || !errors.Is(err, psi.ErrNodeNotFound)) {
		return nil, err
	}

	if resolved != nil {
		return resolved, nil
	}

	root, err := tx.Resolve(ctx, RootPath)

	if err != nil {
		return nil, err
	}

	currentParent := root

	for _, pkg := range pkgs {
		n := currentParent.ResolveChild(ctx, psi.PathElement{Name: pkg})

		if n == nil {
			pkg := NewPackage(pkg)
			pkg.SetParent(currentParent)

			if err := pkg.Update(ctx); err != nil {
				return nil, err
			}

			n = pkg
		}

		currentParent = n
		resolved = n.(*Package)
	}

	return resolved, nil
}
