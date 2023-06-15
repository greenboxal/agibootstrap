package codex

import (
	"context"
	"go/build"
	"go/types"
	"path"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/packages"

	"github.com/greenboxal/agibootstrap/pkg/vts"
)

type AnalysisBuildStep struct{}

type AnalysisContext struct {
	project      *Project
	pkgConfig    *packages.Config
	loaderConfig *loader.Config
	buildContext build.Context

	merr   error
	errors []error
}

func (a *AnalysisBuildStep) Process(ctx context.Context, p *Project) (result BuildStepResult, err error) {
	actx := &AnalysisContext{}

	actx.project = p

	actx.pkgConfig = &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedImports,
		Dir:  p.rootPath,
		Fset: p.fset,
	}

	actx.buildContext = build.Default
	actx.buildContext.Dir = p.rootPath
	actx.buildContext.BuildTags = []string{"selfwip", "psionly"}

	pkgs, err := packages.Load(actx.pkgConfig, path.Join(p.rootPath, "..."))

	if err != nil {
		return result, err
	}

	actx.loaderConfig = &loader.Config{
		Build:       &actx.buildContext,
		Cwd:         p.rootPath,
		AllowErrors: true,
	}

	for _, pkg := range pkgs {
		actx.loaderConfig.Import(pkg.PkgPath)
	}

	program, err := actx.loaderConfig.Load()

	if err != nil {
		return result, err
	}

	for _, pkg := range program.Imported {
		if err := actx.analyzePackage(ctx, pkg); err != nil {
			actx.merr = multierror.Append(actx.merr, err)
		}
	}

	if actx.merr != nil {
		return result, actx.merr
	}

	return
}

func (actx *AnalysisContext) analyzePackage(ctx context.Context, info *loader.PackageInfo) error {
	pkg := &vts.Package{
		Name: vts.PackageName(info.Pkg.Name()),
	}

	for _, typ := range info.Types {
		named, ok := typ.Type.(*types.Named)

		if !ok {
			continue
		}

		pkg.Types = append(pkg.Types,
			&vts.Type{
				Name: vts.TypeName{
					Pkg:  pkg.Name,
					Name: named.Obj().Name(),
				},
			},
		)
	}

	actx.project.vtsRoot.AddPackage(pkg)

	return nil
}
