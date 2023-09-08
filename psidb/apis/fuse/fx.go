package fuse

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"sync"

	fusefs "github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"go.uber.org/fx"

	logging "github.com/greenboxal/agibootstrap/pkg/platform/logging"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
	"github.com/greenboxal/agibootstrap/psidb/psi"
)

var logger = logging.GetLogger("psifuse")

var Module = fx.Module(
	"apis/fuse",

	fx.Provide(NewManager),
)

const FUSE_MOUNT_DIR = "/tmp/psidb/fuse"

type Manager struct {
	m      sync.Mutex
	path   string
	core   coreapi.Core
	server *fuse.Server
}

func NewManager(
	lc fx.Lifecycle,
	core coreapi.Core,
) *Manager {
	m := &Manager{
		path: FUSE_MOUNT_DIR,
		core: core,
	}

	lc.Append(fx.Hook{
		OnStart: m.Start,
		OnStop:  m.Stop,
	})

	return m
}

func (m *Manager) Start(ctx context.Context) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.server != nil {
		return nil
	}

	if err := os.MkdirAll(m.path, 0755); err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		_ = exec.Command("diskutil", "unmount", "force", m.path).Run()
	}

	root := NewPsiNodeDir(m.core, psi.MustParsePath("QmYXZ//"))

	srv, err := fusefs.Mount(m.path, root, &fusefs.Options{})

	if err != nil {
		return nil
	}

	m.server = srv

	return nil
}

func (m *Manager) Stop(ctx context.Context) error {
	m.m.Lock()
	defer m.m.Unlock()

	if m.server != nil {
		if err := m.server.Unmount(); err != nil {
			return err
		}

		m.server = nil
	}

	return nil
}
