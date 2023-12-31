package codex

import (
	"context"
	"fmt"
	"go/build"
	"go/parser"
	"go/types"
	"path"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/packages"

	"github.com/greenboxal/agibootstrap/pkg/codex/vts"
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
	fmt.Println("Analyzing project...")

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
		Fset:        p.fset,
		AllowErrors: true,
		ParserMode:  parser.ParseComments,
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

// analyzePackage analyzes a single Go package and adds it to the VTS root.
// The VTS root tracks all packages, types, functions and other symbols that can be referenced.
func (actx *AnalysisContext) analyzePackage(ctx context.Context, info *loader.PackageInfo) error {
	pkg := &vts.Package{
		Name: vts.PackageName(info.Pkg.Name()),
	}

	for _, typ := range info.Types {
		named, ok := typ.Type.(*types.Named)

		if !ok {
			continue
		}
		typ := &vts.Type{
			Name: vts.TypeName{
				Pkg:  pkg.Name,
				Name: named.Obj().Name(),
			},
		}

		if st, ok := named.Underlying().(*types.Struct); ok {
			// Add fields
			for i := 0; i < st.NumFields(); i++ {
				field := st.Field(i)
				fieldName := field.Name()
				fieldType := field.Type().String()
				f := &vts.Field{
					DeclarationType: vts.TypeName{
						Pkg:  pkg.Name,
						Name: named.Obj().Name(),
					},
					Name: fieldName,
					Type: vts.TypeName{
						Pkg:  pkg.Name,
						Name: fieldType,
					},
				}
				typ.Members = append(typ.Members, f)
			}
		}

		// Add methods
		for i := 0; i < named.NumMethods(); i++ {
			method := named.Method(i)
			methodName := method.Name()
			methodType := method.Type().Underlying().(*types.Signature)

			m := &vts.Method{
				DeclarationType: vts.TypeName{
					Pkg:  pkg.Name,
					Name: named.Obj().Name(),
				},
				Name:           methodName,
				Parameters:     []vts.Parameter{},
				Results:        []vts.Parameter{},
				TypeParameters: []vts.Parameter{},
			}

			// Add parameters
			for j := 0; j < methodType.Params().Len(); j++ {
				param := methodType.Params().At(j)
				paramName := param.Name()
				paramType := param.Type().String()
				p := vts.Parameter{
					Name: paramName,
					Type: vts.TypeName{
						Pkg:  pkg.Name,
						Name: paramType,
					},
				}
				m.Parameters = append(m.Parameters, p)
			}

			// Add results
			for j := 0; j < methodType.Results().Len(); j++ {
				param := methodType.Results().At(j)
				paramType := param.Type().String()
				p := vts.Parameter{
					Type: vts.TypeName{
						Pkg:  pkg.Name,
						Name: paramType,
					},
				}
				m.Results = append(m.Results, p)
			}

			typ.Members = append(typ.Members, m)
		}

		pkg.Types = append(pkg.Types, typ)
	}

	actx.project.vts.AddPackage(pkg)

	return nil
}
