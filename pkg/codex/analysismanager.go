package codex

import (
	"context"

	"github.com/alitto/pond"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type AnalysisManager struct {
	p      *Project
	logger *zap.SugaredLogger

	workerPool *pond.WorkerPool
}

func NewAnalysisManager(p *Project) *AnalysisManager {
	return &AnalysisManager{
		p:      p,
		logger: logging.GetLogger("analysis-manager"),

		workerPool: pond.New(100, 1000),
	}
}

func (am *AnalysisManager) analyzeFile(ctx context.Context, n *vfs.File) error {
	ft := am.p.FileTypeProvider().GetForPath(n.Name)

	if ft == nil {
		return nil
	}

	am.logger.Debugw("analyzing file", "path", n.Path, "filetype", ft)

	langFt, ok := ft.(project.LanguageFileType)

	if !ok {
		return nil
	}

	src, err := project.GetOrCreateSourceForFile(ctx, n, langFt.GetLanguage())

	if err != nil {
		return err
	}

	return src.Load(ctx)
}

func (am *AnalysisManager) Analyze(ctx context.Context, root psi.Node) error {
	grp, gctx := am.workerPool.GroupContext(ctx)

	grp.Submit(func() error {
		return psi.Walk(root, func(cursor psi.Cursor, entering bool) error {
			if !entering {
				return nil
			}

			switch n := cursor.Value().(type) {
			case *vfs.File:
				grp.Submit(func() error {
					return am.analyzeFile(gctx, n)
				})

				cursor.SkipChildren()
				cursor.SkipEdges()

			default:
				cursor.WalkChildren()
				cursor.SkipEdges()
			}

			return nil
		})
	})

	return grp.Wait()
}

func (am *AnalysisManager) Close() error {
	am.workerPool.StopAndWait()

	return nil
}
