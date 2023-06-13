package codex

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/hashicorp/go-multierror"
	"github.com/zeroflucs-given/generics/collections/stack"
	"golang.org/x/exp/slices"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/io"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
)

// A Project is the root of a codex project.
// It contains all the information about the project.
// It is also the entry point for all codex operations.
type Project struct {
	rootPath string
	fs       repofs.FS

	files []string
}

func NewProject(rootPath string) (*Project, error) {
	root, err := repofs.NewFS(rootPath)

	if err != nil {
		return nil, err
	}

	return &Project{
		rootPath: rootPath,
		fs:       root,
	}, nil
}

func (p *Project) RootPath() string { return p.rootPath }

func (p *Project) FS() repofs.FS { return p.fs }

func (p *Project) Sync() error {
	p.files = []string{}

	err := filepath.WalkDir(p.rootPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && isGoFile(path) {
			p.files = append(p.files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *Project) Commit(addAll bool) error {
	isDirty, err := p.fs.IsDirty()

	if err != nil {
		return err
	}

	if !isDirty {
		return nil
	}

	diff, err := p.fs.GetStagedChanges()

	if err != nil {
		return err
	}

	commitMessage, err := gpt.PrepareCommitMessage(diff)

	if err != nil {
		return err
	}

	commitId, err := p.fs.Commit(commitMessage, addAll)

	if err != nil {
		return err
	}

	fmt.Printf("Changes committed with commit ID %s\n", commitId)

	return nil
}

func (p *Project) Generate() (changes int, err error) {
	for {
		stepChanges, err := p.processGenerateStep()

		if err != nil {
			return changes, nil
		}

		if stepChanges == 0 {
			break
		}

		changes += stepChanges

		importsChanges, err := p.processImportsStep()

		if err != nil {
			return changes, nil
		}

		changes += importsChanges

		fixChanges, err := p.processFixStep()

		if err != nil {
			return changes, nil
		}

		changes += fixChanges
	}

	return
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

func (p *Project) processImportsStep() (changes int, err error) {
	for _, file := range p.files {
		opt := &imports.Options{
			FormatOnly: false,
			AllErrors:  true,
			Comments:   true,
			TabIndent:  true,
			TabWidth:   4,
			Fragment:   false,
		}

		code, err := os.ReadFile(file)

		if err != nil {
			return changes, err
		}

		newCode, err := imports.Process(file, code, opt)

		if err != nil {
			return changes, err
		}

		if string(newCode) != string(code) {
			err = io.WriteFile(file, string(newCode))

			if err != nil {
				return changes, err
			}

			changes++
		}
	}

	return
}

// buildProject is responsible for analyzing the project and checking its types.
// It returns a slice of BuildError and an error. BuildError contains information about type-checking errors and their associated package name, filename, line, column and error.
func (p *Project) buildProject() (sf *psi.SourceFile, errs []*BuildError, err error) {
	sf = psi.NewSourceFile("")

	// Get the module path of the package
	modulePath, err := getModulePath(p.rootPath)
	if err != nil {
		return nil, nil, err
	}

	// Get the import path of the package
	packageName := modulePath

	// Set up the build context
	buildContext := build.Default

	// Get all packages in the project
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes | packages.NeedSyntax}, p.rootPath)

	if err != nil {
		return nil, nil, err
	}

	// Iterate through every Go package in the project
	for _, pkg := range pkgs {
		if !pkg.Types.Complete() {
			return nil, nil, fmt.Errorf("incomplete package type info: %q", pkg.ID)
		}

		if pkg.Name == "main" {
			continue // Skip the main package
		}

		if _, ok := pkg.Imports[packageName]; !ok {
			continue // Skip packages that do not import the package we want to analyze
		}

		fset := sf.FileSet()

		// Create the type checker
		typeConfig := &types.Config{
			Error:    func(err error) { /* ignore parse errors */ },
			Importer: p,
			Sizes:    types.SizesFor(buildContext.Compiler, buildContext.GOARCH), // Required for type-checking constants
		}

		// Iterate over each Go source file in the package
		for _, file := range pkg.Syntax {
			// Type-check the file
			info := types.Info{
				Types:      make(map[ast.Expr]types.TypeAndValue),
				Defs:       make(map[*ast.Ident]types.Object),
				Uses:       make(map[*ast.Ident]types.Object),
				Implicits:  make(map[ast.Node]types.Object),
				Selections: make(map[*ast.SelectorExpr]*types.Selection),
			}

			_, err = typeConfig.Check(pkg.ID, fset, []*ast.File{file}, &info)

			if err != nil {
				errs = append(errs, &BuildError{
					Pkg:      pkg,
					Filename: pkg.Fset.File(file.Pos()).Name(),
					Line:     fset.Position(file.Pos()).Line,
					Column:   fset.Position(file.Pos()).Column,
					Error:    err,
				})
			}
		}
	}

	return
}

func (p *Project) processFixStep() (changes int, err error) {
	sf, buildErrors, err := p.buildProject()

	if err != nil {
		return 0, err
	}

	getNodeForError := func(err *BuildError) psi.Node {
		pos := token.Position{Filename: err.Filename, Line: err.Line, Column: err.Column}
		return sf.FindNodeByPos(pos)
	}

	if len(buildErrors) > 0 {
		strs := make([]string, len(buildErrors))
		// Iterate through the build errors and process each error node
		for i, buildError := range buildErrors {
			strs[i] = fmt.Sprintf("%s:%d:%d: %v", buildError.Filename, buildError.Line, buildError.Column, buildError.Error)
			// Replace the error node with the result of processing
			node := getNodeForError(buildError)

			p.ProcessNode(sf, node)

			e := fmt.Errorf("%d errors occurred during type checking in package %s: %v", len(strs), buildError.Pkg.ID, buildErrors)
			err = multierror.Append(err, e)
		}

		return
	}

	return changes, nil
}

// getModulePath returns the module path of the given directory
func getModulePath(dir string) (string, error) {
	cmd := exec.Command("go", "list", "-m", "-json", ".")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	var info struct {
		Path string
	}

	err = json.Unmarshal(out, &info)
	if err != nil {
		return "", err
	}

	return info.Path, nil
}

type BuildError struct {
	Pkg      *packages.Package
	Filename string
	Line     int
	Column   int
	Error    error
}

func (p *Project) processGenerateStep() (changes int, err error) {
	err = filepath.Walk(p.rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".go" {
			count, err := p.processFile(path)

			if err != nil {
				fmt.Printf("Error processing file %v: %v\n", path, err)
				return nil
			}

			changes += count
		}

		return nil
	})

	isDirty, err := p.fs.IsDirty()

	if err != nil {
		return changes, err
	}

	if !isDirty {
		return changes, nil
	}

	err = p.fs.StageAll()

	if err != nil {
		return changes, err
	}

	err = p.Commit(false)

	if err != nil {
		return changes, err
	}

	err = p.fs.Push()

	if err != nil {
		fmt.Printf("Error pushing the changes: %v\n", err)
		return changes, err
	}

	return changes, nil
}

func (p *Project) processFile(fsPath string) (int, error) {
	// Read the file
	code, err := os.ReadFile(fsPath)
	if err != nil {
		return 0, err
	}

	// Parse the file into an AST
	ast := psi.Parse(fsPath, string(code))

	if ast.Error() != nil {
		return 0, err
	}

	// Process the AST nodes
	updated := p.ProcessNodes(ast)

	// Convert the AST back to code
	newCode, err := ast.ToCode(updated)
	if err != nil {
		return 0, err
	}

	// Write the new code to a new file
	err = io.WriteFile(fsPath, newCode)
	if err != nil {
		return 0, err
	}

	if newCode != string(code) {
		return 1, nil
	}

	return 0, nil
}

func (p *Project) ProcessNodes(sf *psi.SourceFile) psi.Node {
	// Process the AST nodes
	updated := p.ProcessNode(sf, sf.Root())

	// Convert the AST back to code
	return updated
}

// ProcessNode processes the given node and returns the updated node.
func (p *Project) ProcessNode(sf *psi.SourceFile, root psi.Node) psi.Node {
	ctx := &NodeProcessor{
		SourceFile:   sf,
		Root:         root,
		FuncStack:    stack.NewStack[*FunctionContext](16),
		Declarations: map[string]*declaration{},
	}

	for _, child := range sf.Root().Children() {
		decl, ok := child.Node().(dst.Decl)

		if !ok {
			continue
		}

		index := slices.Index(ctx.Root.Node().(*dst.File).Decls, decl)

		if index == -1 {
			continue
		}

		names := getDeclarationNames(child)

		for _, name := range names {
			ctx.setExistingDeclaration(index, name, child)
		}
	}

	return psi.Apply(ctx.Root, ctx.OnEnter, ctx.OnLeave)
}

func isGoFile(path string) bool {
	return filepath.Ext(path) == ".go"
}
