package mdutils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/md"
	"github.com/gomarkdown/markdown/parser"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
)

var DefaultTemplateHelpers = map[string]any{
	"json":         Json,
	"markdownTree": MarkdownTree,

	"renderMarkdown": func(level int, input any) string {
		return RenderWithContext(&RenderContext{HeadingLevel: level}, input)
	},
}

func init() {
	RegisterMarkdownHelpers()
}

func RegisterMarkdownHelpers() {
	chain.RegisterDefaultMarkdownHelpers(DefaultTemplateHelpers)
}

func Json(input any) string {
	data, err := json.Marshal(input)

	if err != nil {
		panic(err)
	}

	return string(data)
}
func MarkdownTree(initialDepth int, input any) string {
	var walk func(any, int, string, ast.Node)

	walk = func(node any, depth int, key string, parent ast.Node) {
		val := reflect.ValueOf(node)

		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		switch val.Kind() {
		case reflect.Map:
			for it := val.MapRange(); it.Next(); {
				k := it.Key().String()
				v := it.Value().Interface()

				h := strings.Repeat("#", depth)
				heading := ParseMarkdown([]byte(fmt.Sprintf("%s %s", h, k)))

				walk(v, depth+1, k, heading)

				ast.AppendChild(parent, heading)
			}

		case reflect.Array:
			fallthrough
		case reflect.Slice:
			for i := 0; i < val.Len(); i++ {
				walk(val.Index(i), depth+1, "", parent)
			}

		default:
			leaf := RenderToNode(&RenderContext{HeadingLevel: depth}, node)

			ast.AppendChild(parent, leaf)
		}
	}

	root := &ast.Document{}

	walk(input, initialDepth, "", root)

	str := string(FormatMarkdown(root))

	return str
}

func FormatMarkdown(node ast.Node) []byte {
	return markdown.Render(node, md.NewRenderer())
}

func ParseMarkdown(md []byte) ast.Node {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	return p.Parse(md)
}

func MarkdownToHtml(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	renderer := html.NewRenderer(html.RendererOptions{})

	return markdown.Render(doc, renderer)
}
