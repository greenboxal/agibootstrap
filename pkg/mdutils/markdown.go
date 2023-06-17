package mdutils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
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
	var payload any

	data, err := json.Marshal(input)

	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		panic(err)
	}

	var walk func(any, int, string, ast.Node)

	walk = func(node any, depth int, key string, parent ast.Node) {
		switch node := node.(type) {
		case map[string]any:
			for k, v := range node {
				h := strings.Repeat("#", depth)
				heading := ParseMarkdown([]byte(fmt.Sprintf("%s %s", h, k)))

				walk(v, depth+1, k, heading)

				ast.AppendChild(parent, heading)
			}

		case []any:
			for _, v := range node {
				walk(v, depth+1, "", parent)
			}

		default:
			leaf := &ast.CodeBlock{
				IsFenced: false,
			}

			str := fmt.Sprintf("%s", node)
			lines := strings.Split(str, "\n")
			for i, l := range lines {
				lines[i] = "\t" + l
			}
			str = strings.Join(lines, "\n")

			leaf.Literal = []byte(str)

			ast.AppendChild(parent, leaf)
		}
	}

	root := &ast.Document{}

	walk(payload, initialDepth, "", root)

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
