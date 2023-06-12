package astparser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/zeroflucs-given/generics/collections/stack"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type FunctionContext struct {
	Processor *NodeProcessor
	Func      *ast.FuncDecl
	Todos     []string
}

type NodeProcessor struct {
	SourceFile *psi.SourceFile
	Root       psi.Node
	FuncStack  *stack.Stack[*FunctionContext]
}

func (p *NodeProcessor) OnEnter(cursor *psi.Cursor) bool {
	switch node := cursor.Node().(type) {
	case *ast.FuncDecl:
		err := p.FuncStack.Push(&FunctionContext{
			Func: node,
		})

		if err != nil {
			panic(err)
		}

		for _, cmt := range cursor.Element().Comments() {
			if strings.Contains(cmt.Text, "TODO") {
				ok, currentFn := p.FuncStack.Peek()

				if !ok {
					break
				}

				currentFn.Todos = append(currentFn.Todos, cmt.Text)
			}
		}

	case *ast.Comment:
		if strings.Contains(node.Text, "TODO") {
			ok, currentFn := p.FuncStack.Peek()

			if !ok {
				break
			}

			currentFn.Todos = append(currentFn.Todos, node.Text)
		}

	case *ast.CommentGroup:
		for _, comment := range node.List {
			if strings.Contains(comment.Text, "TODO") {
				ok, currentFn := p.FuncStack.Peek()

				if !ok {
					continue
				}

				currentFn.Todos = append(currentFn.Todos, comment.Text)
			}
		}
	}

	return true
}

func (p *NodeProcessor) OnLeave(cursor *psi.Cursor) bool {
	switch node := cursor.Node().(type) {
	case *ast.FuncDecl:
		ok, currentFn := p.FuncStack.Pop()

		if !ok || len(currentFn.Todos) == 0 {
			return true
		}

		newNode, err := p.Step(currentFn, node)

		if err != nil {
			panic(err)
		}

		if newFunc, ok := newNode.(*ast.FuncDecl); ok {
			cursor.Replace(newNode)

			if f, ok := p.Root.Node().(*ast.File); ok {
				f.Comments = append(f.Comments, newFunc.Doc)
			}
		}

		return true
	}

	return true
}

func (p *NodeProcessor) Step(ctx *FunctionContext, stepRoot ast.Node) (result ast.Node, err error) {
	// Extract the comment
	todoComment := strings.Join(ctx.Todos, "\n")

	prunedRoot := psi.Apply(p.Root, func(cursor *psi.Cursor) bool {
		switch node := cursor.Node().(type) {
		case *ast.FuncDecl:
			if node != ctx.Func {
				cursor.Replace(&ast.FuncDecl{
					Doc:  node.Doc,
					Recv: node.Recv,
					Name: node.Name,
					Type: node.Type,
					Body: &ast.BlockStmt{
						List: []ast.Stmt{},
					},
				})
			}

			return false
		}

		return true
	}, nil)

	// Format Node N to string
	contextStr, err := p.SourceFile.ToCode(prunedRoot)

	if err != nil {
		return nil, err
	}

	stepStr, err := ToCode(stepRoot)

	if err != nil {
		return nil, err
	}

	// Send the string and comment to gpt-3.5-turbo and get a response
	gptResponse, err := gpt.SendToGPT(todoComment, contextStr, stepStr)

	if err != nil {
		return nil, err
	}

	// Parse the generated code into an AST
	newRoot, err := Parse(gptResponse)

	if err != nil {
		previousErr := err
		sourceWithPreamble := fmt.Sprintf("package %s\n\n%s", "gptimport", gptResponse)
		newRoot, err = Parse(sourceWithPreamble)

		if err != nil {
			return nil, previousErr
		}
	}

	rootFile := p.SourceFile.Root().Node().(*ast.File)

	for _, decl := range newRoot.Decls {
		exists := false

		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Name.Name == ctx.Func.Name.Name {
				result = fn
				exists = true
			}
		}

		if !exists {
			for _, d := range rootFile.Decls {
				if d == decl {
					exists = true
					break
				}
			}
		}

		if !exists {
			rootFile.Decls = append(rootFile.Decls, decl)
		}
	}

	return
}

// ProcessNodes processes all AST nodes.
func ProcessNodes(sf *psi.SourceFile) psi.Node {
	ctx := &NodeProcessor{
		SourceFile: sf,
		Root:       sf.Root(),
		FuncStack:  stack.NewStack[*FunctionContext](16),
	}

	return psi.Apply(ctx.Root, ctx.OnEnter, ctx.OnLeave)
}
