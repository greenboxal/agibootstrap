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

var hasPackageRegex = regexp.MustCompile(`(?m)^.*package\s+([a-zA-Z0-9_]+)\n`)
var hasHtmlEscapeRegex = regexp.MustCompile(`&lt;|&gt;|&amp;|&quot;|&#[0-9]{2};`)

// NodeScope represents the context of a function.
//
// The ProcessContext stores information about the processor, node, and todos associated avec with a function.
//
// Fields:
// - Processor: A pointer to the NodeProcessor struct.
// - Node: The psi.Node representing the current function.
// - Todos: A slice of strings representing the todos comments associated with the function.
type NodeScope struct {
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
	Project      *Project                 // The project associated with the NodeProcessor.
	SourceFile   *psi.SourceFile          // The source file being processed.
	Root         psi.Node                 // The root node of the AST being processed.
	FuncStack    *stack.Stack[*NodeScope] // A stack of FunctionContexts.
	Declarations map[string]*declaration  // A map of declaration names to declaration information.

	prepareObjective   func(p *NodeProcessor, ctx *NodeScope) (string, error)                                                 // A function to prepare the objective for GPT-3.
	prepareContext     func(p *NodeProcessor, ctx *NodeScope, root psi.Node, baseRequest gpt.Request) (gpt.ContextBag, error) // A function to prepare the context for GPT-3.
	checkShouldProcess func(fn *NodeScope, cursor *psi.Cursor) bool                                                           // A function to check if a function should be processed.

	ctx    context.Context
	cancel context.CancelFunc
}

// OnEnter method is called when entering a node during
// the AST traversal. It checks if the node is a container,
// and if so, pushes a new NodeScope onto the FuncStack.
// Additionally, it scans the comments of the node for TODOs and stores them in the current NodeScope.
//
// Parameters:
// - cursor: The psi.Cursor representing the current node.
//
// Returns:
// - bool: true to continue traversing the AST, false to stop.
//
// OnEnter is responsible for pushing a new NodeScope onto the FuncStack if the current node is a container.
// Additionally, it scans the comments of the node for TODOs and stores them in the current NodeScope.
func (p *NodeProcessor) OnEnter(cursor *psi.Cursor) bool {
	e := cursor.Element()

	if e.IsContainer() {
		err := p.FuncStack.Push(&NodeScope{
			Processor: p,
			Node:      cursor.Element(),
			Todos:     make([]string, 0),
		})

		if err != nil {
			panic(err)
		}
	}

	for _, txt := range cursor.Element().Comments() {
		txt = strings.TrimSpace(txt)

		if strings.HasPrefix(txt, "// TODO:") {
			ok, currentFn := p.FuncStack.Peek()

			if !ok {
				break
			}

			currentFn.Todos = append(currentFn.Todos, txt)
		}
	}

	return true
}

// OnLeave method is called when leaving a node during
// the AST traversal. It checks if the node is a container,
// and if so, pops the top NodeScope from the FuncStack.
// It also checks if the current function should be processed
// and calls the Step method to process the function if necessary.
//
// Parameters:
// - cursor: The psi.Cursor representing the current node.
//
// Returns:
// - bool: true to continue traversing the AST, false to stop.
//
// OnLeave is responsible for popping the top NodeScope from the FuncStack if the current node is a container.
// Additionally, it checks if the current function should be processed and calls the Step method to process the function if necessary.
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

		_, err := p.Step(p.ctx, currentFn, cursor)

		if err != nil {
			panic(err)
		}

		return false
	}

	return true
}

// The Step function is responsible for code processing and response generation. The algorithm is divided into several groups of steps:
//
// Acquiring and Preparing Elements:
// 1. Retrieve the root element from the cursor using the cursor.Element() method.
// 2. Formulate the todoComment with the p.prepareObjective function, which takes the NodeProcessor and NodeScope as inputs.
//
// Cloning and Pruning:
// 3. Make a clone of the root with the psi.Clone method from the psi package, applied to p.Root. Prune this cloned root by iterating through each cursor element and returning true.
//
// String Conversion and Request Preparation:
// 4. Convert the stepRoot into a string format using the p.SourceFile.ToCode method.
// 5. Construct a new gpt.Request object, where the Document is stepStr, the Objective is todoComment, and the Context is an empty ContextBag.
//
// Context Setup and Invocation:
// 6. Craft the fullContext using the p.prepareContext function with the NodeProcessor, NodeScope, prunedRoot, and req as parameters.
// 7. Assign fullContext as the Context of req.
// 8. Execute the gpt.Invoke function using ctx and req, yielding a codeBlocks response.
//
// Code Blocks Processing:
// 9. Iterate over each codeBlock in codeBlocks, performing the following:
//   - If the Language attribute of the block is empty, set it to "go".
//   - Create a blockName with the fmt.Sprintf function by concatenating "mergeContents_#", i, and block.Language.
//   - Unescape HTML escape sequences in block.Code using the html.UnescapeString function.
//   - Modify the package declaration by enclosing block.Code with "package gptimport".
//   - Parse the modified code into a new Abstract Syntax Tree (AST) using p.SourceFile.Parse, with blockName and patchedCode as inputs.
//   - Merge the newly created AST (newRoot.Ast().(*dst.File)) with the existing AST using the MergeFiles function.
//   - For each declaration (decl) in newRoot.Children(), check if it is a function and if its name matches the current function's name.
//   - If yes, replace the current declaration in the cursor with the new one using p.ReplaceDeclarationAt.
//   - If no, merge the new declaration with the existing declarations using p.MergeDeclarations.
//
// Return Processed Code:
// 10. Return the processed code as a dst.Node.
func (p *NodeProcessor) Step(ctx context.Context, scope *NodeScope, cursor *psi.Cursor) (result dst.Node, err error) {
	stepRoot := cursor.Element()

	todoComment, err := p.prepareObjective(p, scope)
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
		Context:   gpt.ContextBag{},
	}

	fullContext, err := p.prepareContext(p, scope, prunedRoot, req)
	if err != nil {
		return nil, err
	}

	req.Context = fullContext

	codeBlocks, err := gpt.Invoke(ctx, req)
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
			if funcType, ok := decl.Ast().(*dst.FuncDecl); ok && funcType.Name.Name == scope.Node.Ast().(*dst.FuncDecl).Name.Name {
				p.ReplaceDeclarationAt(cursor, decl, funcType.Name.Name)
			} else {
				p.MergeDeclarations(cursor, decl)
			}
		}
	}

	return
}

// MergeDeclarations merges the declarations of a node into the NodeProcessor.
// It takes a psi.Cursor and a psi.Node representing the current node, and returns a boolean value indicating whether the merging was successful or not.
//
// The process of merging declarations involves the following steps:
// 1. Retrieve all declaration names from the node using the getDeclarationNames function.
// 2. For each declaration name, check if it already exists in the NodeProcessor. If not, insert the declaration at the cursor position using the InsertDeclarationAt function.
// 3. If the declaration already exists, check if the current cursor node is the same as the previous node. If so, replace the previous node with the current node's AST representation using the cursor.Replace function.
// 4. Update the existing declaration with the current node's index, name, and AST representation using the setExistingDeclaration function.
//
// The purpose of MergeDeclarations is to ensure that all declarations within a node are properly merged into the NodeProcessor, allowing further processing and code generation to be performed accurately.
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

// InsertDeclarationAt inserts a declaration after the given cursor.
// It takes a psi.Cursor, a name string, and a decl psi.Node.
// The process involves the following steps:
// 1. Calling the InsertAfter method of the cursor and passing decl.Ast() to insert the declaration after the cursor.
// 2. Getting the current index of decl in the root file's declarations using the slices.Index method.
// 3. Calling the setExistingDeclaration method of the NodeProcessor to update the existing declaration information.
//
// The purpose of InsertDeclarationAt is to insert a declaration at a specific position in the AST and update the declaration information in the NodeProcessor for further processing and code generation.
func (p *NodeProcessor) InsertDeclarationAt(cursor *psi.Cursor, name string, decl psi.Node) {
	cursor.InsertAfter(decl.Ast())
	index := slices.Index(p.Root.Ast().(*dst.File).Decls, decl.Ast().(dst.Decl))
	p.setExistingDeclaration(index, name, decl)
}

// ReplaceDeclarationAt method replaces a declaration at a specific cursor position with a new declaration.
//
// The ReplaceDeclarationAt method takes three parameters: a psi.Cursor, the declaration node to replace, and the name of the declaration.
//
// The steps involved in the ReplaceDeclarationAt method are as follows:
// 1. The method replaces the declaration node at the cursor position with the new declaration node using the cursor.Replace method.
// 2. It gets the index of the new declaration in the root file's declarations using the slices.Index method.
// 3. It updates the existing declaration information in the NodeProcessor by calling the setExistingDeclaration method.
//
// The purpose of the ReplaceDeclarationAt method is to provide a mechanism for replacing a declaration at a specific position in the AST and updating the declaration information in the NodeProcessor for further processing and code generation.
func (p *NodeProcessor) ReplaceDeclarationAt(cursor *psi.Cursor, decl psi.Node, name string) {
	cursor.Replace(decl.Ast())
	index := slices.Index(p.Root.Ast().(*dst.File).Decls, decl.Ast().(dst.Decl))
	p.setExistingDeclaration(index, name, decl)
}

// setExistingDeclaration updates the information about an existing declaration.
// It takes an index, a name string, and a psi.Node representing the declaration.
// The process involves the following steps:
// 1. Retrieve the existing declaration from the NodeProcessor using the name.
// 2. If the declaration does not exist, create a new declaration and add it to the NodeProcessor.
// 3. Update the declaration's element, node, and index with the provided values.
//
// This function is responsible for maintaining and updating the information about existing declarations,
// ensuring their accuracy and consistency throughout the code generation process.
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

// getDeclarationNames returns a slice of strings representing the declaration names in the given PSI node.
// The node parameter should be a psi.Node representing the AST node being processed.
// The function iterates through the AST node and extracts the names of declarations such as constants, variables, types, and functions.
// The extracted names are then appended to the names slice and returned.
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
