package codex

import (
	"errors"
	"fmt"
	"go/build"
	"go/types"
	"path"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/greenboxal/agibootstrap/pkg/io"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type BuildError struct {
	Pkg      *packages.Package
	Filename string
	Line     int
	Column   int
	Error    error
}

func (be BuildError) String() string {
	// Implement the BuildError string representation
	return fmt.Sprintf("Package: %s, File: %s, Line: %d, Column: %d, Error: %s", be.Pkg.Name, be.Filename, be.Line, be.Column, be.Error.Error())
}

// FixBuildStep is responsible for fixing all build errors that were found
type FixBuildStep struct{}

func (s *FixBuildStep) Process(p *Project) (result BuildStepResult, err error) {
	buildErrors, err := p.buildProject()

	if err != nil {
		return result, err
	}

	for _, buildError := range buildErrors {
		sf := p.GetSourceFile(buildError.Filename)
		err = s.ProcessFix(p, sf, buildError)

		if err != nil {
			return result, err
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

	// Get all packages in the project
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
		Dir:  p.rootPath,
	}, path.Join(p.rootPath, "..."))

	if err != nil {
		return errs, err
	}

	// Iterate through every Go package in the project
	for _, pkg := range pkgs {
		fmt.Printf("Checking package %s\n", pkg.ID)

		if !pkg.Types.Complete() {
			errs = append(errs, &BuildError{
				Pkg:      pkg,
				Filename: pkg.GoFiles[0],
				Line:     0,
				Column:   0,
				Error:    errors.New("type-checking incomplete"),
			})
			continue
		}

		// Create the type checker
		typeConfig := &types.Config{
			GoVersion: "1.20.3",
			Error:     func(err error) { /* ignore parse errors */ },
			Importer:  &ProjectPackageImporter{Project: p},
			Sizes:     types.SizesFor(buildContext.Compiler, buildContext.GOARCH), // Required for type-checking constants
		}

		// Iterate over each Go source file in the package
		_, err = typeConfig.Check(pkg.ID, pkg.Fset, pkg.Syntax, pkg.TypesInfo)

		if err != nil {
			buildError := &BuildError{
				Pkg:   pkg,
				Error: err,
			}

			if err, ok := err.(types.Error); ok {
				if !strings.HasPrefix(pkg.PkgPath, buildError.Filename) {
					continue
				}

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
}

func (imp *ProjectPackageImporter) Import(path string) (*types.Package, error) {
	// Load the package
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedImports,
		Dir:  imp.Project.RootPath(),
	}, path)

	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, errors.New("unexpected number of packages found")
	}

	return pkgs[0].Types, nil
}
