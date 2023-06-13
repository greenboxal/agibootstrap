package codex

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/types"
	"path"

	"golang.org/x/tools/go/packages"

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

// buildProject is responsible for analyzing the project and checking its types.
// It returns a slice of BuildError and an error. BuildError contains information about type-checking errors and their associated package name, filename, line, column and error.
func (p *Project) buildProject() (errs []*BuildError, err error) {
	// Set up the build context
	buildContext := build.Default

	// Get all packages in the project
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes | packages.NeedSyntax, Dir: p.rootPath}, path.Join(p.rootPath, "..."))

	if err != nil {
		return nil, err
	}

	// Iterate through every Go package in the project
	for _, pkg := range pkgs {
		if !pkg.Types.Complete() {
			return nil, fmt.Errorf("incomplete package type info: %q", pkg.ID)
		}

		// Create the type checker
		typeConfig := &types.Config{
			Error:    func(err error) { /* ignore parse errors */ },
			Importer: p,
			Sizes:    types.SizesFor(buildContext.Compiler, buildContext.GOARCH), // Required for type-checking constants
		}

		// Iterate over each Go source file in the package
		for _, file := range pkg.Syntax {
			_, err = typeConfig.Check(pkg.ID, pkg.Fset, []*ast.File{file}, pkg.TypesInfo)

			if err != nil {
				errs = append(errs, &BuildError{
					Pkg:      pkg,
					Filename: pkg.Fset.File(file.Pos()).Name(),
					Line:     pkg.Fset.Position(file.Pos()).Line,
					Column:   pkg.Fset.Position(file.Pos()).Column,
					Error:    err,
				})
			}
		}
	}

	return
}

// processFixStep is responsible for fixing all build errors that were found
func (p *Project) processFixStep() (changes int, err error) {
	buildErrors, err := p.buildProject()

	if err != nil {
		return 0, err
	}

	if len(buildErrors) > 0 {
		// Iterate through the build errors and process each error node
		for _, buildError := range buildErrors {
			sf := p.sourceFiles[buildError.Filename]
			err = p.ProcessFix(sf, buildError)
			if err != nil {
				return 0, err
			}
			// Increase the count of changes made
			changes++
		}
	}

	return changes, nil
}

// ProcessFix applies a fix to a build error. It takes in a psi.SourceFile pointer and a BuildError pointer and returns an error.
// The function sets the 'prepareObjective' field of the NodeProcessor passed into the p.ProcessNodes function to a function that returns a string that includes the build error message.
// The 'prepareObjective' function is responsible for generating a string that describes what needs to be done to fix a build error.
// The expected input parameters are the psi.SourceFile 'sf' and pointer to the BuildError 'buildError' that needs to be fixed.
// The expected output parameter is an error, which is nil if the process finishes successfully.
func (p *Project) ProcessFix(sf *psi.SourceFile, buildError *BuildError) error {
	p.ProcessNodes(sf, func(p *NodeProcessor) {
		p.prepareObjective = func(p *NodeProcessor, ctx *FunctionContext) (string, error) {
			return "Fix the following build error: " + buildError.Error.Error(), nil
		}
	})

	return nil
}

func (p *Project) Import(path string) (*types.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.LoadTypes,
	}, path)

	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, errors.New("unexpected number of packages found")
	}

	return pkgs[0].Types, nil
}
