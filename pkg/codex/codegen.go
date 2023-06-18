package codex

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/zeroflucs-given/generics/collections/stack"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type CodeGenBuildStep struct{}

func (g *CodeGenBuildStep) Process(ctx context.Context, p *Project) (result BuildStepResult, err error) {
	for _, file := range p.files {
		filePath := file.Path()
		lang := p.langRegistry.ResolveExtension(filePath)

		if lang == nil {
			continue
		}

		count, e := p.processFile(ctx, filePath)

		if e != nil {
			err = multierror.Append(err, e)
		}

		result.Changes += count
	}

	return result, err
}

func (p *Project) processFile(ctx context.Context, fsPath string, opts ...NodeProcessorOption) (int, error) {
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
	updated, err := p.ProcessNodes(ctx, sf, opts...)

	if err != nil {
		return 0, err
	}

	// Convert the AST back to code
	newCode, err := sf.ToCode(updated)
	if err != nil {
		return 0, err
	}

	if newCode.Code != sf.OriginalText() {
		if err := sf.Replace(newCode.Code); err != nil {
			return 0, nil
		}

		return 1, nil
	}

	return 0, nil
}

func (p *Project) ProcessNodes(ctx context.Context, sf psi.SourceFile, opts ...NodeProcessorOption) (psi.Node, error) {
	return p.ProcessNode(ctx, sf, sf.Root(), opts...)
}

// ProcessNode processes the given node and returns the updated node.
func (p *Project) ProcessNode(ctx context.Context, sf psi.SourceFile, root psi.Node, opts ...NodeProcessorOption) (psi.Node, error) {
	processor := &NodeProcessor{
		Project:      p,
		SourceFile:   sf,
		Root:         root,
		FuncStack:    stack.NewStack[*NodeScope](16),
		Declarations: map[string]*declaration{},
	}

	processor.ctx, processor.cancel = context.WithCancel(ctx)

	processor.checkShouldProcess = func(fn *NodeScope, cursor psi.Cursor) bool {
		return len(fn.Todos) > 0
	}

	processor.prepareContext = func(processor *NodeProcessor, ctx *NodeScope, root psi.Node, req gpt.Request) (gpt.ContextBag, error) {
		var err error

		result := gpt.ContextBag{}
		wholeFile, err := processor.SourceFile.ToCode(root)

		if err != nil {
			return nil, err
		}

		hits, err := p.repo.Query(context.Background(), wholeFile.Code, 10)

		if err != nil {
			return nil, err
		}

		for _, hit := range hits {
			key := fmt.Sprintf("for reference only, do not copy: %s @ %d", hit.Entry.Document.Path, hit.Entry.Chunk.Index)

			result[key] = mdutils.CodeBlock{
				Language: "",
				Filename: hit.Entry.Document.Path,
				Code:     hit.Entry.Chunk.Content,
			}
		}

		return result, nil
	}

	processor.prepareObjective = func(p *NodeProcessor, ctx *NodeScope) (string, error) {
		return strings.Join(ctx.Todos, "\n"), nil
	}

	for _, opt := range opts {
		opt(processor)
	}

	result, err := psi.Rewrite(processor.Root, func(cursor psi.Cursor, entering bool) error {
		if entering {
			return processor.OnEnter(cursor)
		} else {
			return processor.OnLeave(cursor)
		}
	})

	if err != nil {
		return nil, err
	}

	result.Update()

	return result, nil
}
