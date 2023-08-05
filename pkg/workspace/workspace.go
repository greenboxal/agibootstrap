package workspace

import (
	"context"
	"os"
	"strings"

	"github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/fti"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

type Workspace struct {
	logger *zap.SugaredLogger

	core coreapi.Core

	repo   *fti.Repository
	vfsm   *vfs.Manager
	rootFs vfs.FileSystem
}

func NewWorkspace(
	lc fx.Lifecycle,
	core coreapi.Core,
	vfsm *vfs.Manager,
) (*Workspace, error) {
	w := &Workspace{
		logger: logging.GetLogger("workspace"),

		core: core,
		vfsm: vfsm,
	}

	root, err := vfsm.CreateLocalFS(core.Config().ProjectDir)

	if err != nil {
		return nil, err
	}

	w.rootFs = root

	repo, err := fti.NewRepository(core.Config().ProjectDir)

	if err != nil {
		return nil, err
	}

	w.repo = repo

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return w.OnStart(ctx)
		},

		OnStop: func(ctx context.Context) error {
			return w.Close(ctx)
		},
	})

	return w, nil
}

func (w *Workspace) OnStart(ctx context.Context) error {
	goprocess.Go(func(p goprocess.Process) {
		ctx := goprocessctx.OnClosingContext(p)

		err := w.core.RunTransaction(ctx, func(ctx context.Context, tx coreapi.Transaction) error {
			root, err := tx.Resolve(ctx, psi.PathFromElements(w.core.Config().RootUUID, false))

			if err != nil {
				return err
			}

			srcsNode := root.ResolveChild(ctx, psi.PathElement{Name: "srcs"})

			if srcsNode == nil {
				srcs := &vfs.Directory{}
				srcs.Name = "srcs"
				srcs.Path = w.core.Config().ProjectDir
				srcs.VFSM = w.vfsm
				srcs.Init(srcs)
				srcs.SetParent(root)

				srcsNode = srcs
			}

			srcsDirectory := srcsNode.(*vfs.Directory)

			return w.performSync(ctx, srcsDirectory)
		})

		if err != nil {
			w.logger.Error(err)
		}
	})

	return nil
}

func (w *Workspace) performSync(ctx context.Context, root psi.Node) error {
	count := 0

	w.logger.Infow("Performing sync walk")

	err := psi.Walk(root, func(cursor psi.Cursor, entering bool) error {
		n := cursor.Value()

		if n, ok := n.(*vfs.Directory); ok && entering {
			count++

			cursor.WalkChildren()

			return n.Sync(func(path string) bool {
				if w.repo.IsIgnored(path) {
					return false
				}

				s, err := os.Stat(path)

				if err != nil {
					return false
				}

				if s.IsDir() {
					return true
				}

				return strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".md")
			})
		} else {
			cursor.SkipChildren()
		}

		return nil
	})

	if err != nil {
		return err
	}

	w.logger.Info("Performing sync update")

	return root.Update(ctx)
}

func (w *Workspace) Close(ctx context.Context) error {
	return nil
}
