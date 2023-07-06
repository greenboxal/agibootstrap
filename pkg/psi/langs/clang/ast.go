package clang

import (
	"context"
	"fmt"

	"github.com/antlr4-go/antlr/v4"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Node interface {
	psi.Node

	Initialize(self Node)

	Ast() antlr.ParserRuleContext
}

type NodeBase[T antlr.ParserRuleContext] struct {
	psi.NodeBase

	node     T
	comments []string
}

func (nb *NodeBase[T]) IsContainer() bool            { return len(nb.Children()) > 0 }
func (nb *NodeBase[T]) IsLeaf() bool                 { return len(nb.Children()) == 0 }
func (nb *NodeBase[T]) Ast() antlr.ParserRuleContext { return nb.node }
func (nb *NodeBase[T]) Comments() []string           { return nb.comments }

func (nb *NodeBase[T]) String() string {
	return fmt.Sprintf("%T(%d, %s)", nb.node, nb.ID(), nb.UUID())
}

func (nb *NodeBase[T]) Initialize(self Node) {
	nb.NodeBase.Init(self)
}

func (nb *NodeBase[T]) OnUpdate(context.Context) error {
	if nb.IsValid() {
		return nil
	}

	nb.NodeBase.OnUpdate(nil)

	return nil
}

func NewNodeFor(node antlr.ParserRuleContext) *NodeBase[antlr.ParserRuleContext] {
	n := &NodeBase[antlr.ParserRuleContext]{node: node}

	n.Initialize(n)

	return n
}

type astConversionContext struct {
	sf *SourceFile

	parentStack []Node
	result      Node
}

func (a *astConversionContext) VisitTerminal(node antlr.TerminalNode) {
}

func (a *astConversionContext) VisitErrorNode(node antlr.ErrorNode) {
}

func (a *astConversionContext) EnterEveryRule(ctx antlr.ParserRuleContext) {
	n := NewNodeFor(ctx)

	if a.sf != nil {
		hidden := a.sf.tokens.GetHiddenTokensToLeft(ctx.GetStart().GetTokenIndex(), 2)

		for _, tk := range hidden {
			if tk.GetChannel() == 2 {
				n.comments = append(n.comments, tk.GetText())
			}
		}
	}

	if len(a.parentStack) > 0 {
		parent := a.parentStack[len(a.parentStack)-1]
		n.SetParent(parent)
	}

	a.parentStack = append(a.parentStack, n)
}

func (a *astConversionContext) ExitEveryRule(ctx antlr.ParserRuleContext) {
	a.result = a.parentStack[len(a.parentStack)-1]
	a.parentStack = a.parentStack[:len(a.parentStack)-1]
}

func AstToPsi(sf *SourceFile, parsed antlr.ParserRuleContext) psi.Node {
	ctx := &astConversionContext{
		sf: sf,
	}

	walker := antlr.NewParseTreeWalker()
	walker.Walk(ctx, parsed)

	return ctx.result
}
