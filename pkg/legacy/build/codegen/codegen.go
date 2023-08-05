package codegen

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	build2 "github.com/greenboxal/agibootstrap/pkg/legacy/build"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

type BuildStep struct{}

func (bs *BuildStep) Process(ctx context.Context, bctx *build2.Context) (result build2.StepResult, err error) {
	langRegistry := bctx.Project().LanguageProvider()

	err = psi.Walk(bctx.Project(), func(cursor psi.Cursor, entering bool) error {
		n := cursor.Value()

		if entering {
			switch n := n.(type) {
			case project.Project:
				cursor.WalkChildren()

			case *vfs.Directory:
				cursor.WalkChildren()

			case *vfs.File:
				filePath := n.GetPath()
				lang := langRegistry.ResolveExtension(filePath)

				if lang == nil {
					break
				}

				count, e := bs.processFile(ctx, bctx, filePath)

				if e != nil {
					err = multierror.Append(err, e)
				}

				result.ChangeCount += count
				cursor.SkipChildren()

			default:
				cursor.SkipChildren()
			}
		}

		return nil
	})

	return
}

func (bs *BuildStep) processFile(ctx context.Context, bctx *build2.Context, fsPath string, opts ...NodeProcessorOption) (int, error) {
	p := bctx.Project()

	//bctx.Branch().Infow("Processing file", "file", fsPath)

	sf, err := p.GetSourceFile(ctx, fsPath)

	if err != nil {
		return 0, err
	}

	if sf.Error() != nil {
		return 0, err
	}

	// Process the AST nodes
	updated, err := bs.ProcessNode(ctx, bctx, sf, sf.Root(), opts...)

	if err != nil {
		return 0, err
	}

	// Convert the AST back to code
	newCode, err := sf.ToCode(updated)
	if err != nil {
		return 0, err
	}

	if newCode.Code != sf.OriginalText() {
		if err := sf.Replace(ctx, newCode.Code); err != nil {
			return 0, nil
		}

		return 1, nil
	}

	return 0, nil
}

// ProcessNode processes the given node and returns the updated node.
func (bs *BuildStep) ProcessNode(ctx context.Context, bctx *build2.Context, sf project.SourceFile, root psi.Node, opts ...NodeProcessorOption) (psi.Node, error) {
	processor := &NodeProcessor{
		Project:    bctx.Project(),
		SourceFile: sf,
		Root:       root,
	}

	processor.ctx, processor.cancel = context.WithCancel(ctx)

	processor.checkShouldProcess = func(fn *NodeScope, cursor psi.Cursor) bool {
		return len(fn.Todos) > 0
	}

	processor.prepareContext = func(processor *NodeProcessor, ctx *NodeScope, root psi.Node, req gpt.CodeGeneratorRequest) (gpt.ContextBag, error) {
		var err error

		result := gpt.ContextBag{}

		wholeFile, err := processor.SourceFile.ToCode(root)

		if err != nil {
			return nil, err
		}

		queries := []string{wholeFile.Code}

		if req.Objective != "" {
			queries = append(queries, req.Objective)
		}

		if req.Plan != "" {
			queries = append(queries, req.Plan)
		}

		for _, query := range queries {
			hits, err := bctx.Project().Repo().Query(context.Background(), query, 5)

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
		}

		return result, nil
	}

	processor.prepareObjective = func(p *NodeProcessor, ctx *NodeScope) (result string, err error) {
		for _, todo := range ctx.Todos {
			result += fmt.Sprintf("- [ ] %s\n", todoRegex.ReplaceAllString(todo, ""))
		}

		return
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

	if err := result.Update(ctx); err != nil {
		return nil, err
	}

	return result, nil
}
