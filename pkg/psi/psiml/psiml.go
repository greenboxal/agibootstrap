package psiml

import (
	"bytes"
	"context"

	"github.com/go-yaml/yaml"
	"github.com/gomarkdown/markdown/ast"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering/themes"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
	"github.com/greenboxal/agibootstrap/psidb/db/online"
	"github.com/greenboxal/agibootstrap/psidb/modules/stdlib"
	"github.com/greenboxal/agibootstrap/psidb/services"
)

type TextProcessor struct {
	lg     *online.LiveGraph
	search *services.SearchService
}

func NewTextProcessor(
	lg *online.LiveGraph,
	search *services.SearchService,

) *TextProcessor {
	return &TextProcessor{lg: lg, search: search}
}

func (tp *TextProcessor) Process(ctx context.Context, text string) (_ string, parseErr error) {
	parsed := mdutils.ParseMarkdown([]byte(text))

	ast.WalkFunc(parsed, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch node := node.(type) {
			case *ast.CodeBlock:
				parent := node.Parent.AsContainer()
				index := slices.Index(parent.Children, ast.Node(node))

				result, err := tp.renderNode(ctx, node)

				if err != nil {
					parseErr = multierror.Append(parseErr, err)
					return ast.Terminate
				}

				ast.RemoveFromTree(node)

				parent.Children = slices.Insert(parent.Children, index, result)
				result.SetParent(parent)
			}
		}

		return ast.GoToNext
	})

	if parseErr != nil {
		return "", parseErr
	}

	return string(mdutils.FormatMarkdown(parsed)), nil
}

func (tp *TextProcessor) renderNode(ctx context.Context, node *ast.CodeBlock) (ast.Node, error) {
	lang := string(node.Info)
	content := string(node.Literal)

	switch lang {
	case "psi-query":
		return tp.renderPsi(ctx, content)
	}

	return node, nil
}

type Query struct {
	From  psi.Path `json:"from"`
	Query string   `json:"query"`
	Limit int      `json:"limit"`

	View string `json:"view"`
}

func (tp *TextProcessor) renderPsi(ctx context.Context, content string) (ast.Node, error) {
	var q Query

	if err := yaml.Unmarshal([]byte(content), &q); err != nil {
		return nil, err
	}

	if q.Limit == 0 {
		q.Limit = 10
	}

	result, err := tp.search.Search(ctx, &services.SearchRequest{
		Graph: tp.lg,
		Query: stdlib.NewText(q.Query),
		Limit: q.Limit,
		Scope: q.From,

		ReturnNode: true,
	})

	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer

	for _, hit := range result.Results {
		if err := rendering.RenderNodeWithTheme(ctx, &buffer, themes.GlobalTheme, "text/markdown", q.View, hit.Node); err != nil {
			return nil, err
		}
	}

	node := mdutils.ParseMarkdown(buffer.Bytes())

	return node, nil
}
