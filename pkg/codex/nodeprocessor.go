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

// FunctionContext represents the context of a function.
//
// The ProcessContext stores information about the processor, node, and todos associated avec with a function.
//
// Fields:
// - Processor: A pointer to the NodeProcessor struct.
// - Node: The psi.Node representing the current function.
// - Todos: A slice of strings representing the todos comments associated with the function.
type FunctionContext struct {
	Processor *NodeProcessor
	Node      psi.Node
	Todos     []string
}

// declaration represents a declaration in the code.
//
// It contains information about the declaration such as the AST node, the PSI node, the index, and the name.
//
// Fields:
// - node: The AST node representing the declaration.
// - element: The PSI node representing the declaration.
// - index: The index of the declaration.
// - name: The name of the declaration.
type declaration struct {
	node    dst.Node
	element psi.Node
	index   int
	name    string
}

// NodeProcessorOption is a function type that defines an option for the NodeProcessor.
//
// It is used to configure the behavior of the NodeProcessor. Each option is a function that takes
// a pointer to the NodeProcessor as a parameter and modifies its properties in some way.
type NodeProcessorOption func(p *NodeProcessor)

// NodeProcessor is responsible for processing AST nodes and generating code.
// It is used to configure the behavior of the NodeProcessor. Each option is a function that takes
// a pointer to the NodeProcessor as a parameter and modifies its properties in some way.
type NodeProcessor struct {
	Project      *Project                       // The project associated with the NodeProcessor.
	SourceFile   *psi.SourceFile                // The source file being processed.
	Root         psi.Node                       // The root node of the AST being processed.
	FuncStack    *stack.Stack[*FunctionContext] // A stack of FunctionContexts.
	Declarations map[string]*declaration        // A map of declaration names to declaration information.

	prepareObjective   func(p *NodeProcessor, ctx *FunctionContext) (string, error)                                                 // A function to prepare the objective for GPT-3.
	prepareContext     func(p *NodeProcessor, ctx *FunctionContext, root psi.Node, baseRequest gpt.Request) (gpt.ContextBag, error) // A function to prepare the context for GPT-3.
	checkShouldProcess func(fn *FunctionContext, cursor *psi.Cursor) bool                                                           // A function to check if a function should be processed.
}

// OnEnter method is called when entering a node during
// the AST traversal. It checks if the node is a container,
// and if so, pushes a new FunctionContext onto the FuncStack.
// Additionally, it scans the comments of the node for TODOs and stores them in the current FunctionContext.
//
// Parameters:
// - cursor: The psi.Cursor representing the current node.
//
// Returns:
// - bool: true to continue traversing the AST, false to stop.
//
// OnEnter is responsible for pushing a new FunctionContext onto the FuncStack if the current node is a container.
// Additionally, it scans the comments of the node for TODOs and stores them in the current FunctionContext.
func (p *NodeProcessor) OnEnter(cursor *psi.Cursor) bool {
	e := cursor.Element()

	if e.IsContainer() {
		err := p.FuncStack.Push(&FunctionContext{
			Processor: p,
			Node:      cursor.Element(),
			Todos:     make([]string, 0),
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

func orphanSnippet0() {
	e := cursor.Element()

	if e.IsContainer() {
		err := p.FuncStack.Push(&FunctionContext{
			Processor: p,
			Node:      cursor.Element(),
			Todos:     make([]string, 0),
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

// TODO: Write documentation explaining the process, the steps involved and its purpose.
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

// Step processes the code and generates a response
func (p *NodeProcessor) Step(ctx *FunctionContext, cursor *psi.Cursor) (result dst.Node, err error) {
	stepRoot := cursor.Element()

	todoComment, err := p.prepareObjective(p, ctx)

	if err != nil {
		return nil, err
	}

	prunedRoot := psi.Apply(psi.Clone(p.Root), func(cursor *psi.Cursor) bool {
		return true
	}, nil)

	stepStr, err := p.SourceFile.ToCode(stepRoot)

	if err != nil {
		return nil, err
	}

	req := gpt.Request{
		Document:  stepStr,
		Objective: todoComment,

		Context: gpt.ContextBag{},
	}

	if err != nil {
		return nil, err
	}

	// Format Node N to string
	fullContext, err := p.prepareContext(p, ctx, prunedRoot, req)

	if err != nil {
		return nil, err
	}

	req.Context = fullContext

	// Send the string and comment to gpt-3.5-turbo and get a response
	codeBlocks, err := gpt.Invoke(context.TODO(), req)

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

		MergeFiles(newRoot.Ast().(*dst.File), newRoot.Ast().(*dst.File))

		for _, decl := range newRoot.Children() {
			if funcType, ok := decl.Ast().(*dst.FuncDecl); ok && funcType.Name.Name == ctx.Node.Ast().(*dst.FuncDecl).Name.Name {
				p.ReplaceDeclarationAt(cursor, decl, funcType.Name.Name)
			} else {
				p.MergeDeclarations(cursor, decl)
			}
		}
	}

	return
}

func MergeFiles(file1, file2 *dst.File) *dst.File {
	mergedFile := file1
	newDecls := make([]dst.Decl, 0)

	for _, decl2 := range file2.Decls {
		found := false
		switch decl2 := decl2.(type) {
		case *dst.FuncDecl:
			for i, decl1 := range mergedFile.Decls {
				if decl1, ok := decl1.(*dst.FuncDecl); ok && decl1.Name.Name == decl2.Name.Name {
					mergedFile.Decls[i] = decl2
					found = true
					break
				}
			}
		case *dst.GenDecl:
			for _, spec2 := range decl2.Specs {
				switch spec2 := spec2.(type) {
				case *dst.TypeSpec:
					for i, decl1 := range mergedFile.Decls {
						if decl1, ok := decl1.(*dst.GenDecl); ok {
							for j, spec1 := range decl1.Specs {
								if spec1, ok := spec1.(*dst.TypeSpec); ok && spec1.Name.Name == spec2.Name.Name {
									decl1.Specs[j] = spec2
									found = true
									break
								}
							}
						}
						mergedFile.Decls[i] = decl1
					}
				case *dst.ValueSpec:
					for i, decl1 := range mergedFile.Decls {
						if decl1, ok := decl1.(*dst.GenDecl); ok {
							for j, spec1 := range decl1.Specs {
								if spec1, ok := spec1.(*dst.ValueSpec); ok && spec1.Names[0].Name == spec2.Names[0].Name {
									decl1.Specs[j] = spec2
									found = true
									break
								}
							}
						}
						mergedFile.Decls[i] = decl1
					}
				}
			}
		}

		if !found {
			newDecls = append(newDecls, decl2)
		}
	}

	mergedFile.Decls = append(mergedFile.Decls, newDecls...)

	return mergedFile
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

// TODO: Write documentation explaining the process, the steps involved and its purpose.
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

// TODO: Write documentation explaining the process, the steps involved and its purpose.
func (p *NodeProcessor) InsertDeclarationAt(cursor *psi.Cursor, name string, decl psi.Node) {
	cursor.InsertAfter(decl.Ast())
	index := slices.Index(p.Root.Ast().(*dst.File).Decls, decl.Ast().(dst.Decl))
	p.setExistingDeclaration(index, name, decl)
}

// TODO: Write documentation explaining the process, the steps involved and its purpose.
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
