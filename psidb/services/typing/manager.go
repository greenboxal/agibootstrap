package typing

import (
	"context"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
	"go.uber.org/fx"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type Manager struct {
	core     coreapi.Core
	migrator migrations.Migrator

	intrinsicRegistry psi.TypeRegistry

	typeCache  map[typesystem.Type]psi.Path
	jsonSchema jsonschema.Schema
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
		typeCache:         make(map[typesystem.Type]psi.Path),
	}

	m.jsonSchema.Definitions = make(map[string]*jsonschema.Schema)

	lc.Append(fx.Hook{
		OnStart: m.Start,
	})

	return m
}

func (m *Manager) FullJsonSchema() *jsonschema.Schema {
	return &m.jsonSchema
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

func (m *Manager) newTypeFromTypeWithName(ctx context.Context, name string, nt typesystem.Type) (*Type, error) {
	name = strings.ReplaceAll(name, "[", "_QZQZ_")
	name = strings.ReplaceAll(name, "]", "_QZQZ_")
	lastIndex := strings.LastIndex(name, "/")

	t := &Type{
		Name:          name[lastIndex+1:],
		FullName:      name,
		PrimitiveKind: nt.PrimitiveKind(),
		Schema:        nt.JsonSchema(),
	}

	t.Init(t, psi.WithNodeType(TypeType))

	return t, nil
}

func (m *Manager) registerType(ctx context.Context, name typesystem.TypeName, typ typesystem.Type, nt psi.NodeType) (*Type, error) {
	if typ == nil {
		panic("type is nil")
	}
	for typ.RuntimeType().Kind() == reflect.Ptr {
		typ = typesystem.TypeFrom(typ.RuntimeType().Elem())
	}

	if path, ok := m.typeCache[typ]; ok {
		return psi.Resolve[*Type](ctx, coreapi.GetTransaction(ctx).Graph(), path)
	}

	pkg, err := m.lookupPackage(ctx, name.Package, true)

	if err != nil {
		return nil, err
	}

	tn := pkg.ResolveChild(ctx, psi.PathElement{Name: name.NameWithArgs()})

	if tn == nil {
		t, err := m.newTypeFromTypeWithName(ctx, name.NameWithArgs(), typ)

		if err != nil {
			return nil, err
		}

		if nt != nil {
			for _, iface := range nt.Interfaces() {
				def := InterfaceDefinition{
					Name:        iface.Name(),
					Description: iface.Description(),
				}

				for _, action := range iface.Interface().Actions() {
					ad := ActionDefinition{
						Name:        action.Name,
						Description: action.Description,
					}

					if action.RequestType != nil {
						typ, err := m.registerType(ctx, action.RequestType.Name(), action.RequestType, nil)

						if err != nil {
							return nil, err
						}

						p := typ.CanonicalPath()
						ad.RequestType = &p
					}

					if action.ResponseType != nil {
						typ, err := m.registerType(ctx, action.ResponseType.Name(), action.ResponseType, nil)

						if err != nil {
							return nil, err
						}

						p := typ.CanonicalPath()
						ad.ResponseType = &p
					}

					def.Actions = append(def.Actions, ad)
				}

				t.Interfaces = append(t.Interfaces, def)
			}
		}

		t.SetParent(pkg)

		if err := t.Update(ctx); err != nil {
			return nil, err
		}

		tn = t
	}

	registeredType := tn.(*Type)

	m.typeCache[typ] = registeredType.CanonicalPath()
	m.jsonSchema.Definitions[name.FullNameWithArgs()] = registeredType.Schema

	return registeredType, nil
}

func (m *Manager) registerNodeType(ctx context.Context, nt psi.NodeType) (*Type, error) {
	return m.registerType(ctx, nt.TypeName(), nt.Type(), nt)
}

func (m *Manager) CreateType(ctx context.Context, typ *Type) (*Type, error) {
	pkgComponents := typesystem.ParseTypeName(typ.FullName)
	pkgName := pkgComponents.Package

	pkg, err := m.lookupPackage(ctx, pkgName, true)

	if err != nil {
		return nil, err
	}

	typ.SetParent(pkg)

	if err := pkg.Update(ctx); err != nil {
		return nil, err
	}

	return typ, nil
}

func (m *Manager) LookupType(ctx context.Context, name string) (resolved *Type, err error) {
	components := typesystem.ParseTypeName(name)
	pkgName := components.Package
	typeName := components.NameWithArgs()

	pkg, err := m.lookupPackage(ctx, pkgName, false)

	if err != nil {
		return nil, err
	}

	tn := pkg.ResolveChild(ctx, psi.PathElement{Name: typeName})

	if tn == nil {
		return nil, nil
	}

	return tn.(*Type), nil
}

func (m *Manager) lookupPackage(ctx context.Context, name string, create bool) (resolved *Package, err error) {
	tx := coreapi.GetTransaction(ctx)

	pkgs := strings.Split(name, "/")

	p := RootPath

	for _, pkg := range pkgs {
		p = p.Child(psi.PathElement{Name: pkg})
	}

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

			if err := currentParent.Update(ctx); err != nil {
				return nil, err
			}

			n = pkg
		}

		currentParent = n
		resolved = n.(*Package)
	}

	return resolved, nil
}
