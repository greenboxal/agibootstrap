package build

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/project"
)

type Builder struct {
	project project.Project
	cfg     Configuration
}

type Result struct {
	ChangeCount int
	Errors      []error

	TotalSteps  int
	TotalEpochs int
}

func NewBuilder(project project.Project, cfg Configuration) *Builder {
	return &Builder{
		project: project,
		cfg:     cfg,
	}
}

func (b *Builder) Build(ctx context.Context) (*Result, error) {
	var err error

	bctx := &Context{
		builder: b,
		project: b.project,
		cfg:     b.cfg,
	}

	defer func() {
		bctx.Close()
	}()

	bctx.log, err = NewLog(b.cfg.ResolveBuildFile("build.log"))

	if err != nil {
		return nil, err
	}

	return bctx.Build(ctx)
}
