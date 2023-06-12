package codex

import (
	"fmt"
	"strings"

	"github.com/dave/dst"
	"github.com/zeroflucs-given/generics/collections/stack"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type FunctionContext struct {
	Processor *NodeProcessor
	Func      *dst.FuncDecl
	Todos     []string
}

type NodeProcessor struct {
	SourceFile *psi.SourceFile
	Root       psi.Node
	FuncStack  *stack.Stack[*FunctionContext]
}

func (p *NodeProcessor) OnEnter(cursor *psi.Cursor) bool {
	switch node := cursor.Node().(type) {
	case *dst.FuncDecl:
		err := p.FuncStack.Push(&FunctionContext{
			Func: node,
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
	switch cursor.Node().(type) {
	case *dst.FuncDecl:
		ok, currentFn := p.FuncStack.Pop()

		if !ok || len(currentFn.Todos) == 0 {
			return true
		}

		newNode, err := p.Step(currentFn, cursor.Element())

		if err != nil {
			panic(err)
		}

		if _, ok := newNode.(*dst.FuncDecl); ok {
			cursor.Replace(newNode)
		}

		return true
	}

	return true
}

func (p *NodeProcessor) Step(ctx *FunctionContext, stepRoot psi.Node) (result dst.Node, err error) {
	// Extract the comment
	todoComment := strings.Join(ctx.Todos, "\n")

	prunedRoot := psi.Apply(psi.Clone(p.Root), func(cursor *psi.Cursor) bool {
		switch node := cursor.Node().(type) {
		case *dst.FuncDecl:
			if node != ctx.Func {
				cursor.Replace(&dst.FuncDecl{
					Decs: node.Decs,
					Recv: node.Recv,
					Name: node.Name,
					Type: node.Type,
					Body: &dst.BlockStmt{
						List: []dst.Stmt{},
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

	stepStr, err := p.SourceFile.ToCode(stepRoot)

	if err != nil {
		return nil, err
	}

	// Send the string and comment to gpt-3.5-turbo and get a response
	gptResponse, err := gpt.SendToGPT(todoComment, contextStr, stepStr)

	if err != nil {
		return nil, err
	}

	// Parse the generated code into an AST
	newRoot, err := p.SourceFile.Parse("_mergeContents.go", gptResponse)

	if err != nil {
		previousErr := err
		sourceWithPreamble := fmt.Sprintf("package %s\n\n%s", "gptimport", gptResponse)
		newRoot, err = p.SourceFile.Parse("_mergeContents.go", sourceWithPreamble)

		if err != nil {
			return nil, previousErr
		}
	}

	rootFile := p.SourceFile.Root().Node().(*dst.File)

	for _, decl := range newRoot.Node().(*dst.File).Decls {
		if fn, ok := decl.(*dst.FuncDecl); ok {
			if fn.Name.Name == ctx.Func.Name.Name {
				result = fn
			}
		} else {
			idx := -1

			for i, d := range rootFile.Decls {
				if d == decl {
					idx = i
					break
				}
			}

			if idx != -1 {
				rootFile.Decls[idx] = decl
			} else {
				rootFile.Decls = append(rootFile.Decls, decl)
			}
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
