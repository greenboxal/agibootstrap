package codex

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dave/dst"
	"github.com/zeroflucs-given/generics/collections/stack"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/io"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

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
	ast, err := psi.Parse(fsPath, string(code))

	if err != nil {
		return 0, err
	}

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

	if ctx.SourceFile == nil {
		panic("SourceFile is nil")
	}

	if ctx.SourceFile.Root() == nil {
		panic("SourceFile.Root() is nil")
	}

	if ctx.SourceFile.Root().Node() == nil {
		panic("SourceFile.Root().Node() is nil")
	}

	rootFile := ctx.SourceFile.Root().Node().(*dst.File)

	for _, child := range ctx.Root.Children() {
		decl, ok := child.Node().(dst.Decl)

		if !ok {
			continue
		}

		index := slices.Index(rootFile.Decls, decl)

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
