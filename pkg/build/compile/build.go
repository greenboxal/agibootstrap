package compile

import (
	"context"
	"fmt"
	"go/build"
	"go/token"
	"go/types"
	"path"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/packages"

	"github.com/greenboxal/agibootstrap/pkg/codex"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type BuildError struct {
	Pkg      string
	Filename string
	Line     int
	Column   int
	Error    error
}

// Implement the String() method for BuildError
func (be BuildError) String() string {
	return fmt.Sprintf("Package: %s, File: %s, Line: %d, Column: %d, Error: %s", be.Pkg, be.Filename, be.Line, be.Column, be.Error.Error())
}

// FixBuildStep is responsible for fixing all build errors that were found
type FixBuildStep struct{}

func (s *FixBuildStep) Process(ctx context.Context, p *codex.Project) (result codex.BuildStepResult, err error) {
	buildErrors, err := p.buildProject()

	if err != nil {
		return result, err
	}

	for _, buildError := range buildErrors {
		sf, e := p.GetSourceFile(buildError.Filename)

		if e != nil {
			err = multierror.Append(err, e)
			continue
		}

		e = s.ProcessFix(ctx, p, sf, buildError)

		if err != nil {
			err = multierror.Append(err, e)
		}
	}

	return result, nil
}

// ProcessFix applies a fix to a build error. It takes in a psi.SourceFile pointer and a BuildError pointer and returns an error.
// The function sets the 'prepareObjective' field of the NodeProcessor passed into the p.ProcessNodes function to a function that returns a string that includes the build error message.
// The 'prepareObjective' function is responsible for generating a string that describes what needs to be done to fix a build error.
// The expected input parameters are the psi.SourceFile 'sf' and pointer to the BuildError 'buildError' that needs to be fixed.
// The expected output parameter is an error, which is nil if the process finishes successfully.
func (s *FixBuildStep) ProcessFix(ctx context.Context, p *codex.Project, sf psi.SourceFile, buildError *BuildError) error {
	fmt.Printf("Fixing build error: %s\n", buildError.String())

	updated, err := p.ProcessNodes(ctx, sf, func(p *NodeProcessor) {
		p.prepareObjective = func(p *NodeProcessor, ctx *NodeScope) (string, error) {
			return "Fix the following build error:\n\n# Error Log\n```log\n" + buildError.String() + "```", nil
		}

		p.checkShouldProcess = func(fn *NodeScope, cursor psi.Cursor) bool {
			return true
		}
	})

	if err != nil {
		return err
	}

	// Convert the AST back to code
	newCode, err := sf.ToCode(updated)
	if err != nil {
		return err
	}

	return sf.Replace(newCode.Code)
}

// buildProject is responsible for analyzing the project and checking its types.
// It returns a slice of BuildError and an error. BuildError contains information about type-checking errors and their associated package name, filename, line, column and error.
func (p *codex.Project) buildProject() (errs []*BuildError, err error) {
	// Set up the build context
	buildContext := build.Default
	buildContext.Dir = p.rootPath
	buildContext.BuildTags = []string{"selfwip", "psionly"}

	fset := token.NewFileSet()

	pkgConfig := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedImports,
		Dir:  p.rootPath,
		Fset: fset,
	}

	// Get all packages in the project
	pkgs, err := packages.Load(pkgConfig, path.Join(p.rootPath, "..."))

	if err != nil {
		return errs, err
	}

	lconf := loader.Config{
		Build:       &buildContext,
		Cwd:         p.rootPath,
		ImportPkgs:  map[string]bool{},
		AllowErrors: true,
	}

	for _, pkg := range pkgs {
		lconf.Import(pkg.PkgPath)
	}

	pro, err := lconf.Load()

	if err != nil {
		return errs, err
	}

	// Iterate through every Go package in the project
	for _, pkg := range pro.Imported {
		fmt.Printf("Checking package %s\n", pkg.Pkg.Name())

		for _, err := range pkg.Errors {
			buildError := &BuildError{
				Pkg:   pkg.Pkg.Path(),
				Error: err,
			}

			if err, ok := err.(types.Error); ok {
				buildError.Filename = err.Fset.File(err.Pos).Name()
				buildError.Line = err.Fset.File(err.Pos).Line(err.Pos)
				buildError.Column = err.Fset.File(err.Pos).Offset(err.Pos)
			}

			errs = append(errs, buildError)
		}
	}

	return errs, nil
}
