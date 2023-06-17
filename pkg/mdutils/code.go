package mdutils

import "github.com/gomarkdown/markdown/ast"

// CodeBlock represents a block of code with its language and code content.
type CodeBlock struct {
	Filename string
	Language string
	Code     string
}

func (c CodeBlock) Render(ctx *RenderContext) ast.Node {
	h := ctx.SpawnHeading(c.Filename)

	code := &ast.CodeBlock{}
	code.Literal = []byte(c.Code)
	code.Info = []byte(c.Language)
	code.IsFenced = true

	ast.AppendChild(h, code)

	return h
}

// ExtractCodeBlocks traverses the given AST and extracts all code blocks.
// It returns a slice of CodeBlock objects, each representing a code block
// with its language and code content.
func ExtractCodeBlocks(root ast.Node) (blocks []CodeBlock) {
	ast.WalkFunc(root, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch node := node.(type) {
			case *ast.CodeBlock:
				blocks = append(blocks, CodeBlock{
					Language: string(node.Info),
					Code:     string(node.Literal),
				})
			}
		}

		return ast.GoToNext
	})

	return
}
