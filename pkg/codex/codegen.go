package codex

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dave/dst"
	"github.com/hashicorp/go-multierror"
	"github.com/zeroflucs-given/generics/collections/stack"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type CodeGenBuildStep struct{}

func (g *CodeGenBuildStep) Process(p *Project) (result BuildStepResult, err error) {
	for _, file := range p.files {
		filePath := file.Path
		if filepath.Ext(filePath) == ".go" {
			count, e := p.processFile(filePath)

			if e != nil {
				err = multierror.Append(err, e)
			}

			result.Changes += count
		}
	}

	return result, err
}

func (p *Project) processFile(fsPath string, opts ...NodeProcessorOption) (int, error) {
	fmt.Printf("Processing file %s\n", fsPath)

	sf, err := p.GetSourceFile(fsPath)

	if err != nil {
		return 0, err
	}

	if sf.Error() != nil {
		return 0, err
	}

	p.sourceFiles[fsPath] = sf

	// Process the AST nodes
	updated := p.ProcessNodes(sf, opts...)

	// Convert the AST back to code
	newCode, err := sf.ToCode(updated)
	if err != nil {
		return 0, err
	}

	if newCode != sf.OriginalText() {
		if err := sf.Replace(newCode); err != nil {
			return 0, nil
		}

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
	//buildContext := build.Default
	//buildContext.Dir = p.rootPath
	//buildContext.BuildTags = []string{"selfwip", "psionly"}

	//lconf := loader.Config{
	//	Build:       &buildContext,
	//	Cwd:         p.rootPath,
	//	AllowErrors: true,
	//}

	//lconf.Import(filepath.Dir(sf.Path()))

	//pro, _ := lconf.Load()

	ctx := &NodeProcessor{
		Project:      p,
		SourceFile:   sf,
		Root:         root,
		FuncStack:    stack.NewStack[*FunctionContext](16),
		Declarations: map[string]*declaration{},
	}

	ctx.checkShouldProcess = func(fn *FunctionContext, cursor *psi.Cursor) bool {
		return len(fn.Todos) > 0
	}

	ctx.prepareContext = func(processor *NodeProcessor, ctx *FunctionContext, root psi.Node, baseRequest string) (any, error) {
		var err error

		result := gpt.ContextBag{}

		result["hits"], err = p.repo.Query(context.Background(), baseRequest, 3)

		if err != nil {
			return "", err
		}

		result["file"], err = processor.SourceFile.ToCode(root)

		if err != nil {
			return "", err
		}

		return result, nil
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
		panic("SourceFile.Root().Ast() is nil")
	}

	rootFile := ctx.SourceFile.Root().Ast().(*dst.File)

	for _, child := range ctx.Root.Children() {
		decl, ok := child.Ast().(dst.Decl)

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
