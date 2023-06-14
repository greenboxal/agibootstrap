package codex

import (
	"context"
	"fmt"
	"go/scanner"
	"html"
	"regexp"
	"strings"

	"github.com/dave/dst"
	"github.com/hashicorp/go-multierror"
	"github.com/zeroflucs-given/generics/collections/stack"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type FunctionContext struct {
	Processor *NodeProcessor
	Node      psi.Node
	Todos     []string
}

type declaration struct {
	node    dst.Node
	element psi.Node
	index   int
	name    string
}

type NodeProcessorOption func(p *NodeProcessor)

type NodeProcessor struct {
	SourceFile   *psi.SourceFile
	Root         psi.Node
	FuncStack    *stack.Stack[*FunctionContext]
	Declarations map[string]*declaration

	prepareObjective   func(p *NodeProcessor, ctx *FunctionContext) (string, error)
	prepareContext     func(p *NodeProcessor, ctx *FunctionContext, root psi.Node) (string, error)
	checkShouldProcess func(fn *FunctionContext, cursor *psi.Cursor) bool
}

func (p *NodeProcessor) OnEnter(cursor *psi.Cursor) bool {
	e := cursor.Element()

	if e.IsContainer() {
		err := p.FuncStack.Push(&FunctionContext{
			Node: cursor.Element(),
		})

		if err != nil {
			panic(err)
		}
	}

	for _, txt := range cursor.Element().Comments() {
		if strings.Contains(txt, "TODO") {
			ok, currentFn := p.FuncStack.Peek()

			if !ok {
				break
			}

			currentFn.Todos = append(currentFn.Todos, txt)
		}
	}

	return true
}

func (p *NodeProcessor) OnLeave(cursor *psi.Cursor) bool {
	e := cursor.Element()

	if e.IsContainer() {
		ok, currentFn := p.FuncStack.Pop()

		if !ok {
			return true
		}

		if !p.checkShouldProcess(currentFn, cursor) {
			return true
		}

		_, err := p.Step(currentFn, cursor)

		if err != nil {
			panic(err)
		}

		return false
	}

	return true
}

var hasPackageRegex = regexp.MustCompile(`(?m)^.*package\s+([a-zA-Z0-9_]+)\n`)
var hasHtmlEscapeRegex = regexp.MustCompile(`&lt;|&gt;|&amp;|&quot;|&#[0-9]{2};`)

func (p *NodeProcessor) Step(ctx *FunctionContext, cursor *psi.Cursor) (result dst.Node, err error) {
	stepRoot := cursor.Element()

	todoComment, err := p.prepareObjective(p, ctx)

	if err != nil {
		return nil, err
	}

	prunedRoot := psi.Apply(psi.Clone(p.Root), func(cursor *psi.Cursor) bool {
		return true
	}, nil)

	// Format Node N to string
	contextStr, err := p.prepareContext(p, ctx, prunedRoot)

	if err != nil {
		return nil, err
	}

	stepStr, err := p.SourceFile.ToCode(stepRoot)

	if err != nil {
		return nil, err
	}

	// Send the string and comment to gpt-3.5-turbo and get a response
	codeBlocks, err := gpt.Invoke(context.TODO(), gpt.Request{
		Document:  stepStr,
		Objective: todoComment,

		Context: gpt.ContextBag{
			"outer_context": contextStr,
		},
	})

	if err != nil {
		return nil, err
	}

	for i, block := range codeBlocks {
		if block.Language == "" {
			block.Language = "go"
		}

		blockName := fmt.Sprintf("_mergeContents_%d.%s", i, block.Language)

		if hasHtmlEscapeRegex.MatchString(block.Code) {
			block.Code = html.UnescapeString(block.Code)
		}

		patchedCode := block.Code
		pkgIndex := hasPackageRegex.FindStringIndex(patchedCode)

		if len(pkgIndex) > 0 {
			patchedCode = fmt.Sprintf("%s%s%s", patchedCode[:pkgIndex[1]], "\n", patchedCode[pkgIndex[1]:])
		} else {
			patchedCode = fmt.Sprintf("package gptimport\n%s", patchedCode)
		}

		patchedCode = hasPackageRegex.ReplaceAllString(patchedCode, "package gptimport\n")

		// Parse the generated code into an AST
		newRoot, e := p.SourceFile.Parse(blockName, patchedCode)

		if e != nil {
			if errList, ok := e.(scanner.ErrorList); ok {
				if len(errList) == 1 && strings.HasPrefix(errList[0].Msg, "expected declaration, ") {
					patchedCode = fmt.Sprintf("package gptimport_orphan\nfunc orphanSnippet%d() {\n%s\n}\n", i, block.Code)
					newRoot2, e2 := p.SourceFile.Parse(blockName, patchedCode)

					if e2 != nil {
						err = multierror.Append(err, e)
						continue
					}

					newRoot = newRoot2
				}
			} else if e != nil {
				err = multierror.Append(err, e)
				continue
			}
		}

		for _, decl := range newRoot.Children() {
			switch n := decl.Ast().(type) {
			case *dst.FuncDecl:
				p.ReplaceDeclarationAt(cursor, decl, n.Name.Name)

			case *dst.GenDecl:
				p.MergeDeclarations(cursor, decl)
			}
		}
	}

	return
}

func (p *NodeProcessor) setExistingDeclaration(index int, name string, node psi.Node) {
	decl := p.Declarations[name]

	if decl == nil {
		decl = &declaration{
			name:    name,
			node:    node.Ast(),
			element: node,
			index:   index,
		}

		p.Declarations[name] = decl
	}

	decl.element = node
	decl.node = node.Ast()
	decl.index = index
}

func (p *NodeProcessor) MergeDeclarations(cursor *psi.Cursor, node psi.Node) bool {
	names := getDeclarationNames(node)

	for _, name := range names {
		previous := p.Declarations[name]

		if previous == nil {
			p.InsertDeclarationAt(cursor, name, node)
		} else {
			if cursor.Node() == previous.node {
				cursor.Replace(node.Ast())
			}

			p.setExistingDeclaration(previous.index, name, node)
		}
	}

	return true
}

func (p *NodeProcessor) InsertDeclarationAt(cursor *psi.Cursor, name string, decl psi.Node) {
	cursor.InsertAfter(decl.Ast())
	index := slices.Index(p.Root.Ast().(*dst.File).Decls, decl.Ast().(dst.Decl))
	p.setExistingDeclaration(index, name, decl)
}

func (p *NodeProcessor) ReplaceDeclarationAt(cursor *psi.Cursor, decl psi.Node, name string) {
	cursor.Replace(decl.Ast())
	index := slices.Index(p.Root.Ast().(*dst.File).Decls, decl.Ast().(dst.Decl))
	p.setExistingDeclaration(index, name, decl)
}

func getDeclarationNames(node psi.Node) []string {
	var names []string

	switch d := node.Ast().(type) {
	case *dst.GenDecl:
		for _, spec := range d.Specs {

			switch s := spec.(type) {
			case *dst.ValueSpec: // for constants and variables
				for _, name := range s.Names {
					names = append(names, name.Name)
				}
			case *dst.TypeSpec: // for types
				names = append(names, s.Name.Name)
			}
		}
	case *dst.FuncDecl: // for functions
		names = append(names, d.Name.Name)
	}

	return names
}
