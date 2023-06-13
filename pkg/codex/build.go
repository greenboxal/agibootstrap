package codex

import (
	"errors"
	"fmt"
	"go/build"
	"go/token"
	"go/types"
	"path"
	"strings"

	"github.com/hashicorp/go-multierror"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/packages"

	"github.com/greenboxal/agibootstrap/pkg/io"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type BuildError struct {
	Pkg      string
	Filename string
	Line     int
	Column   int
	Error    error
}

func (be BuildError) String() string {
	// Implement the BuildError string representation
	return fmt.Sprintf("Package: %s, File: %s, Line: %d, Column: %d, Error: %s", be.Pkg, be.Filename, be.Line, be.Column, be.Error.Error())
}

// FixBuildStep is responsible for fixing all build errors that were found
type FixBuildStep struct{}

func (s *FixBuildStep) Process(p *Project) (result BuildStepResult, err error) {
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

		e = s.ProcessFix(p, sf, buildError)

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
func (s *FixBuildStep) ProcessFix(p *Project, sf *psi.SourceFile, buildError *BuildError) error {
	fmt.Printf("Fixing build error: %s\n", buildError.String())

	updated := p.ProcessNodes(sf, func(p *NodeProcessor) {
		p.prepareObjective = func(p *NodeProcessor, ctx *FunctionContext) (string, error) {
			return "Fix the following build error: " + buildError.String(), nil
		}

		p.checkShouldProcess = func(fn *FunctionContext, cursor *psi.Cursor) bool {
			return true
		}
	})

	// Convert the AST back to code
	newCode, err := sf.ToCode(updated)
	if err != nil {
		return err
	}

	// Write the new code to a new file
	err = io.WriteFile(sf.Path(), newCode)
	if err != nil {
		return err
	}

	return nil
}

// buildProject is responsible for analyzing the project and checking its types.
// It returns a slice of BuildError and an error. BuildError contains information about type-checking errors and their associated package name, filename, line, column and error.
func (p *Project) buildProject() (errs []*BuildError, err error) {
	// Set up the build context
	buildContext := build.Default
	buildContext.Dir = p.rootPath

	fset := token.NewFileSet()
	pkgConfig := &packages.Config{
		BuildFlags: []string{"-modfile", path.Join(p.rootPath, "go.mod"), "-mod=readonly"},
		Mode:       packages.NeedName | packages.NeedFiles | packages.NeedTypes | packages.NeedSyntax | packages.NeedImports,
		Dir:        p.rootPath,
		Fset:       fset,
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
	pro = pro

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

type ProjectPackageImporter struct {
	Project *Project
	Config  *packages.Config
	Loader  *loader.Config
}

func (imp *ProjectPackageImporter) Import(filePath string) (*types.Package, error) {
	if strings.HasPrefix(filePath, "github.com/greenboxal/aip/") {
		filePath = strings.Replace(filePath, "github.com/greenboxal/aip/", "/Users/jonathanlima/IdeaProjects/aip/", 1)
	}

	// Load the package
	pkgs, err := packages.Load(imp.Config, filePath)

	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, errors.New("unexpected number of packages found")
	}

	return pkgs[0].Types, nil
}
