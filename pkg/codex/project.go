package codex

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/zeroflucs-given/generics/collections/stack"
	"golang.org/x/exp/slices"

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
	fs, err := repofs.NewFS(rootPath)

	if err != nil {
		return nil, err
	}

	return &Project{
		rootPath: rootPath,
		fs:       fs,
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

		fixChanges, err := p.processFixStep()

		if err != nil {
			return changes, nil
		}

		changes += fixChanges
	}

	return
}

func (p *Project) processFixStep() (changes int, err error) {
	// Execute goimports
	for _, file := range p.files {
		cmd := exec.Command("goimports", "-w", file)
		err := cmd.Run()
		if err != nil {
			return changes, err
		}
	}

	// Build the project
	cfg := &packages.Config{Mode: packages.LoadSyntax}
	pkgs, err := packages.Load(cfg, fmt.Sprintf("./%s/...", filepath.Base(p.rootPath)))
	if err != nil {
		return changes, err
	}

	// Check for build errors and warnings
	var errorBuffer bytes.Buffer
	var warningBuffer bytes.Buffer
	_, _ = fmt.Fprintln(&errorBuffer, "Build Errors:")
	_, _ = fmt.Fprintln(&warningBuffer, "Build Warnings:")
	var hasErrors bool
	for _, pkg := range pkgs {
		for _, err := range pkg.Errors {
			isBuildTagError := false
			for _, tag := range err.Tags {
				if tag == buildtag.Bad {
					isBuildTagError = true
					break
				}
			}
			if !isBuildTagError {
				hasErrors = true
				_, _ = fmt.Fprintln(&errorBuffer, err.Error())
			}
		}
		for _, warn := range pkg.GoFilesWarned {
			_, _ = fmt.Fprintln(&warningBuffer, warn.Error())
		}
	}

	if hasErrors {
		return changes, errors.New(errorBuffer.String())
	}

	if warningBuffer.Len() > 0 {
		fmt.Println(warningBuffer.String())
	}

	// Re-format files
	for _, file := range p.files {
		err := formatFile(file)
		if err != nil {
			return changes, err
		}
	}

	return len(p.files), nil
}

// formatFile reformats a Go file using gofmt.
func formatFile(file string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	ast.SortImports(fset, node)
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	err = format.Node(f, fset, node)
	if err != nil {
		return err
	}
	return nil
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

// ProcessNodes processes all AST nodes.
func (p *Project) ProcessNodes(sf *psi.SourceFile) psi.Node {
	ctx := &NodeProcessor{
		SourceFile:   sf,
		Root:         sf.Root(),
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

func isGoFile(path string) bool {
	return filepath.Ext(path) == ".go"
}
