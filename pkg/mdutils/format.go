package mdutils

import (
	"fmt"

	"github.com/gomarkdown/markdown/ast"
)

type Renderable interface {
	Render(ctx *RenderContext) ast.Node
}

type RenderableFunc func() ast.Node

type RenderContext struct {
	HeadingLevel int
}

func (ctx *RenderContext) SpawnHeading(title string) *ast.Heading {
	h := &ast.Heading{}
	h.Literal = []byte(title)
	h.Level = ctx.HeadingLevel

	return h
}

func RenderWithContext(ctx *RenderContext, r any) string {
	var node ast.Node

	switch r := r.(type) {
	case Renderable:
		node = r.Render(ctx)

	case ast.Node:
		node = r

	default:
		h := &ast.Heading{}
		h.Literal = []byte(fmt.Sprintf("%T", r))
		h.Level = ctx.HeadingLevel

		txt := &ast.Text{}
		txt.Literal = []byte(MarkdownTree(ctx.HeadingLevel+1, r))
	}

	return string(FormatMarkdown(node))
}
