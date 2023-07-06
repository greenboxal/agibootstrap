package codegen

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/dave/dst"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

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

func (n *NodeScope) Root() psi.Node {
	return n.Node
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
	Project    project.Project // The project associated with the NodeProcessor.
	SourceFile psi.SourceFile  // The source file being processed.
	Root       psi.Node        // The root node of the AST being processed.
	FuncStack  []*NodeScope    // A stack of FunctionContexts.

	prepareObjective   func(p *NodeProcessor, ctx *NodeScope) (string, error)                                                              // A function to prepare the objective for GPT-3.
	prepareContext     func(p *NodeProcessor, ctx *NodeScope, root psi.Node, baseRequest gpt.CodeGeneratorRequest) (gpt.ContextBag, error) // A function to prepare the context for GPT-3.
	checkShouldProcess func(fn *NodeScope, cursor psi.Cursor) bool                                                                         // A function to check if a function should be processed.

	ctx    context.Context
	cancel context.CancelFunc
}

var todoRegex = regexp.MustCompile(`(?m)^\s*//\s*TODO:`)

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
func (p *NodeProcessor) OnEnter(cursor psi.Cursor) error {
	e := cursor.Value()

	if e.IsContainer() {
		scope := &NodeScope{
			Processor: p,
			Node:      cursor.Value(),
			Todos:     make([]string, 0),
		}

		p.FuncStack = append(p.FuncStack, scope)
	}

	for _, txt := range cursor.Value().Comments() {
		txt = strings.TrimSpace(txt)

		if todoRegex.MatchString(txt) && len(p.FuncStack) > 0 {
			currentFn := p.FuncStack[len(p.FuncStack)-1]

			currentFn.Todos = append(currentFn.Todos, txt)
		}
	}

	return nil
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
func (p *NodeProcessor) OnLeave(cursor psi.Cursor) error {
	e := cursor.Value()

	if e.IsContainer() {
		if len(p.FuncStack) == 0 {
			return nil
		}

		currentFn := p.FuncStack[len(p.FuncStack)-1]
		p.FuncStack = p.FuncStack[:len(p.FuncStack)-1]

		if !p.checkShouldProcess(currentFn, cursor) {
			return nil
		}

		_, err := p.Step(p.ctx, currentFn, cursor)

		if err != nil {
			return err
		}

		return psi.ErrAbort
	}

	return nil
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
// String Conversion and CodeGeneratorRequest Preparation:
// 4. Convert the stepRoot into a string format using the p.SourceFile.ToCode method.
// 5. Construct a new gpt.CodeGeneratorRequest object, where the Document is stepStr, the Objective is todoComment, and the Context is an empty ContextBag.
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
func (p *NodeProcessor) Step(ctx context.Context, scope *NodeScope, cursor psi.Cursor) (result dst.Node, err error) {
	stepRoot := cursor.Value()

	todoComment, err := p.prepareObjective(p, scope)
	if err != nil {
		return nil, err
	}

	prunedRoot := p.Root

	rootStr, err := p.SourceFile.ToCode(prunedRoot)
	if err != nil {
		return nil, err
	}

	stepStr, err := p.SourceFile.ToCode(stepRoot)
	if err != nil {
		return nil, err
	}

	req := gpt.CodeGeneratorRequest{
		Document:  rootStr,
		Focus:     stepStr,
		Objective: todoComment,
		Language:  string(p.SourceFile.Language().Name()),
		Context:   gpt.ContextBag{},

		RetrieveContext: func(ctx context.Context, req gpt.CodeGeneratorRequest) (gpt.ContextBag, error) {
			return p.prepareContext(p, scope, p.Root, req)
		},
	}

	fullContext, err := p.prepareContext(p, scope, prunedRoot, req)
	if err != nil {
		return nil, err
	}

	req.Context = fullContext

	cg := gpt.NewCodeGenerator()
	res, err := cg.Generate(ctx, req)

	if err != nil {
		return nil, err
	}

	for i, block := range res.CodeBlocks {
		block.Language = string(p.SourceFile.Language().Name())

		lang := p.Project.LanguageProvider().Resolve(psi.LanguageID(block.Language))

		blockName := fmt.Sprintf("_mergeContents_%d.%s", i, block.Language)

		newRoot, err := lang.ParseCodeBlock(ctx, blockName, block)

		if err != nil {
			return nil, err
		}

		// Printer go brrrrrrrrr
		level := 0
		_ = psi.Walk(newRoot, func(cursor psi.Cursor, entering bool) error {
			if entering {
				level++
			} else {
				level--
			}

			n := cursor.Value()

			fmt.Printf("%s%s\n", strings.Repeat(" ", level*2), n.CanonicalPath())

			return nil
		})

		err = p.SourceFile.MergeCompletionResults(ctx, scope, cursor, newRoot, newRoot.Root())

		if err != nil {
			return nil, err
		}
	}

	return
}
