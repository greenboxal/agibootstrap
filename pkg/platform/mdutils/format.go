package mdutils

import (
	"fmt"
	"reflect"

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

func RenderToNode(ctx *RenderContext, r any) ast.Node {
	var node ast.Node

	switch r := r.(type) {
	case Renderable:
		node = r.Render(ctx)

	case ast.Node:
		node = r

	default:
		v := reflect.ValueOf(r)

		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		h := &ast.Heading{}
		h.Literal = []byte(fmt.Sprintf("%T", r))
		h.Level = ctx.HeadingLevel

		txt := &ast.Text{}

		switch v.Kind() {
		case reflect.Map:
			fallthrough
		case reflect.Slice:
			txt.Literal = []byte(MarkdownTree(ctx.HeadingLevel+1, r))
		default:
			txt.Literal = []byte(fmt.Sprintf("%s", r))
		}
	}

	return node
}

func RenderWithContext(ctx *RenderContext, r any) string {
	node := RenderToNode(ctx, r)

	return string(FormatMarkdown(node))
}
