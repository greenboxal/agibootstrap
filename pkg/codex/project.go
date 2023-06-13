package codex

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/zeroflucs-given/generics/collections/stack"
	"golang.org/x/exp/slices"
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

		fixChanges, err := p.processFixStep()

		if err != nil {
			return changes, nil
		}

		changes += fixChanges
	}

	return
}

func (p *Project) processFixStep() (changes int, err error) {
	for _, file := range p.files {
		opt := &amp
		imports.Options{
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

	build := exec.Command("go", "build")
	out, err := build.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running go build: %s", out)
	}
	// Collect all the errors

	return changes, nil
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

func isGoFile(path string) bool {
	return filepath.Ext(path) == ".go"
}
