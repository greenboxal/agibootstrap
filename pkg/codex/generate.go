package codex

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/hashicorp/go-multierror"
	"github.com/zeroflucs-given/generics/collections/stack"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/io"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func (p *Project) Generate() (changes int, err error) {
	for {
		stepChanges, err := p.processGenerateStep()

		if err != nil {
			return changes, err
		}

		importsChanges, err := p.processImportsStep()

		if err != nil {
			return changes, err
		}

		fixChanges, err := p.processFixStep()

		if err != nil {
			return changes, err
		}

		stepChanges += importsChanges
		stepChanges += fixChanges
		changes += stepChanges

		if stepChanges == 0 {
			break
		}
	}

	return
}

func (p *Project) processGenerateStep() (changes int, err error) {
	for _, path := range p.files {
		if filepath.Ext(path) == ".go" {
			count, e := p.processFile(path)

			if e != nil {
				err = multierror.Append(err, e)
			}

			changes += count
		}
	}

	if err != nil {
		return
	}

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

func (p *Project) processFile(fsPath string, opts ...NodeProcessorOption) (int, error) {
	fmt.Printf("Processing file %s\n", fsPath)

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

	p.sourceFiles[fsPath] = ast

	// Process the AST nodes
	updated := p.ProcessNodes(ast, opts...)

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

func (p *Project) ProcessNodes(sf *psi.SourceFile, opts ...NodeProcessorOption) psi.Node {
	// Process the AST nodes
	updated := p.ProcessNode(sf, sf.Root(), opts...)

	// Convert the AST back to code
	return updated
}

// ProcessNode processes the given node and returns the updated node.
func (p *Project) ProcessNode(sf *psi.SourceFile, root psi.Node, opts ...NodeProcessorOption) psi.Node {
	ctx := &NodeProcessor{
		SourceFile:   sf,
		Root:         root,
		FuncStack:    stack.NewStack[*FunctionContext](16),
		Declarations: map[string]*declaration{},
	}

	ctx.checkShouldProcess = func(fn *FunctionContext, cursor *psi.Cursor) bool {
		return len(fn.Todos) > 0
	}

	ctx.prepareContext = func(p *NodeProcessor, ctx *FunctionContext, root psi.Node) (string, error) {
		return p.SourceFile.ToCode(root)
	}

	ctx.prepareObjective = func(p *NodeProcessor, ctx *FunctionContext) (string, error) {
		return strings.Join(ctx.Todos, "\n"), nil
	}

	for _, opt := range opts {
		opt(ctx)
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
