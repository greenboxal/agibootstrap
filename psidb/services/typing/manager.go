package typing

import (
	"context"
	"reflect"
	"strings"

	"github.com/iancoleman/orderedmap"
	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/typesystem"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/services/migrations"
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
	name = strings.ReplaceAll(name, "/", ".")
	name = strings.ReplaceAll(name, "[", "___")
	name = strings.ReplaceAll(name, "]", "___")
	lastIndex := strings.LastIndex(name, ".")

	t := &Type{
		Name:          name[lastIndex+1:],
		PrimitiveKind: nt.PrimitiveKind(),
	}

	switch nt.PrimitiveKind() {
	case typesystem.PrimitiveKindBoolean:
		t.Schema.Type = "boolean"
	case typesystem.PrimitiveKindUnsignedInt:
		t.Schema.Type = "number"
	case typesystem.PrimitiveKindInt:
		t.Schema.Type = "number"
	case typesystem.PrimitiveKindFloat:
		t.Schema.Type = "number"
	case typesystem.PrimitiveKindStruct:
		t.Schema.Type = "object"
	case typesystem.PrimitiveKindList:
		t.Schema.Type = "array"
	case typesystem.PrimitiveKindString:
		t.Schema.Type = "string"
	case typesystem.PrimitiveKindBytes:
		t.Schema.Type = "string"
	}

	if nt.PrimitiveKind() == typesystem.PrimitiveKindStruct {
		st := nt.Struct()

		t.Schema.Properties = orderedmap.New()

		for i := 0; i < st.NumField(); i++ {
			field := st.FieldByIndex(i)

			typ, err := m.registerType(ctx, field.Type().Name(), field.Type(), nil)

			if err != nil {
				return nil, err
			}

			t.Fields = append(t.Fields, FieldDefinition{
				Name: field.Name(),
				Type: typ.CanonicalPath(),
			})

			t.Schema.Properties.Set(field.Name(), &jsonschema.Schema{
				Ref: "#/$defs/" + field.Type().Name().FullNameWithArgs(),
			})
			t.Schema.Required = append(t.Schema.Required, field.Name())
		}
	} else if nt.PrimitiveKind() == typesystem.PrimitiveKindList {
		inner, err := m.registerType(ctx, nt.List().Elem().Name(), nt.List().Elem(), nil)

		if err != nil {
			return nil, err
		}

		t.Schema.Items = &inner.Schema
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
					Name: iface.Name(),
				}

				for _, action := range iface.Interface().Actions() {
					ad := ActionDefinition{Name: action.Name}

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
	m.jsonSchema.Definitions[name.FullNameWithArgs()] = &registeredType.Schema

	return registeredType, nil
}

func (m *Manager) registerNodeType(ctx context.Context, nt psi.NodeType) (*Type, error) {
	return m.registerType(ctx, nt.TypeName(), nt.Type(), nt)
}

func (m *Manager) lookupType(ctx context.Context, name string) (resolved *Type, err error) {
	pkgComponents := strings.Split(name, ".")
	pkgComponents = pkgComponents[:len(pkgComponents)-1]
	pkgName := strings.Join(pkgComponents, ".")

	pkg, err := m.lookupPackage(ctx, pkgName, false)

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

	pkgs := strings.Split(strings.ReplaceAll(name, "/", "."), ".")

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

func ConvertTypeNameToPath(typeName typesystem.TypeName) psi.Path {
	return ConvertNameToPath(typeName.FullNameWithArgs())
}

func ConvertNameToPath(name string) psi.Path {
	pkgs := strings.Split(name, ".")

	elements := lo.Map(pkgs, func(pkg string, _ int) psi.PathElement {
		return psi.PathElement{Name: pkg}
	})

	return RootPath.Join(psi.PathFromElements("", false, elements...))
}
