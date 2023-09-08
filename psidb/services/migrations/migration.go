package migrations

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"
	"golang.org/x/exp/slices"

	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type MigrationHookFunc func(ctx context.Context, tx coreapi.Transaction) error

type MigrationRecord struct {
	psi.NodeBase

	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

var MigrationRecordType = psi.DefineNodeType[*MigrationRecord]()

func NewMigrationRecord(name string) *MigrationRecord {
	mr := &MigrationRecord{
		Name:      name,
		Timestamp: time.Now(),
	}

	mr.Init(mr)

	return mr
}

func (mr *MigrationRecord) PsiNodeName() string { return mr.Name }

type Migration struct {
	Name string

	Up   MigrationHookFunc
	Down MigrationHookFunc
}

type MigrationSet struct {
	Name       string
	Migrations []Migration
}

func NewNamedMigrationSet(name string, migrations ...Migration) *MigrationSet {
	ms := &MigrationSet{
		Name:       name,
		Migrations: migrations,
	}

	for i, m := range ms.Migrations {
		if m.Name == "" {
			m.Name = fmt.Sprintf("%d---", i)
		}
	}

	slices.SortFunc(ms.Migrations, func(i, j Migration) bool {
		return i.Name < j.Name
	})

	return ms
}

func NewOrderedMigrationSet(name string, migrations ...Migration) *MigrationSet {
	for i, m := range migrations {
		m.Name = fmt.Sprintf("%d-%s", i, m.Name)
	}

	return NewNamedMigrationSet(name, migrations...)
}

func NewMigrationSet(name string, migrations ...Migration) *MigrationSet {
	return &MigrationSet{
		Name:       name,
		Migrations: migrations,
	}
}

type Migrator interface {
	Migrate(ctx context.Context, ms *MigrationSet) error
}

type Manager struct {
	core coreapi.Core
}

func NewManager(lc fx.Lifecycle, core coreapi.Core) *Manager {
	m := &Manager{
		core: core,
	}

	lc.Append(fx.Hook{
		OnStart: m.Start,
	})

	return m
}

func (m *Manager) Start(ctx context.Context) error {
	path := psi.MustParsePath("//_Migrations/")

	return m.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		records, err := psi.ResolveOrCreate[*stdlib.Collection](ctx, tx.Graph(), path, func() *stdlib.Collection {
			return stdlib.NewCollection("_Migrations")
		})

		if err != nil {
			return err
		}

		return records.Update(ctx)
	})
}

func (m *Manager) Migrate(ctx context.Context, ms *MigrationSet) error {
	basePath := psi.MustParsePath("//_Migrations/" + ms.Name)

	return m.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
		records, err := psi.ResolveOrCreate[*stdlib.Collection](ctx, tx.Graph(), basePath, func() *stdlib.Collection {
			return stdlib.NewCollection(ms.Name)
		})

		if err != nil {
			return nil
		}

		for _, migration := range ms.Migrations {
			existing := records.ResolveChild(ctx, psi.PathElement{Name: migration.Name})

			if existing != nil {
				continue
			}

			if err := migration.Up(ctx, tx); err != nil {
				return err
			}

			record := NewMigrationRecord(migration.Name)
			record.SetParent(records)

			if err := record.Update(ctx); err != nil {
				return err
			}
		}

		return nil
	})
}
