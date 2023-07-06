package pylang

import (
	"context"
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Node interface {
	psi.Node

	Initialize(self Node)

	Tree() antlr.ParseTree
	Ast() antlr.ParserRuleContext
	Token() antlr.Token
}

type NodeBase[T antlr.ParseTree] struct {
	psi.NodeBase

	node     T
	comments []string

	sf         *SourceFile
	start, end antlr.Token
	isTerminal bool
}

func (nb *NodeBase[T]) String() string {
	return fmt.Sprintf("%T(%d, %s)", nb.node, nb.ID(), nb.UUID())
}

func (nb *NodeBase[T]) IsContainer() bool  { return !nb.isTerminal }
func (nb *NodeBase[T]) IsLeaf() bool       { return nb.isTerminal }
func (nb *NodeBase[T]) Comments() []string { return nb.comments }

func (nb *NodeBase[T]) Initialize(self Node) {
	nb.NodeBase.Init(self)
}

func (nb *NodeBase[T]) Tree() antlr.ParseTree { return nb.node }

func (nb *NodeBase[T]) Token() antlr.Token {
	if n, ok := any(nb.node).(antlr.TerminalNode); ok {
		return n.GetSymbol()
	}

	return nil
}

func (nb *NodeBase[T]) Ast() antlr.ParserRuleContext {
	if n, ok := any(nb.node).(antlr.ParserRuleContext); ok {
		return n
	}

	return nil
}

func (nb *NodeBase[T]) OnUpdate(context.Context) error {
	if nb.IsValid() {
		return nil
	}

	nb.NodeBase.OnUpdate(nil)

	if nb.IsContainer() {
		for i := 0; i < nb.node.GetChildCount(); i++ {
			nb.Ast().RemoveLastChild()
		}

		for _, c := range nb.Children() {
			cn := c.(Node)

			if c.IsContainer() {
				nb.Ast().AddChild(cn.Ast())
			} else {
				nb.Ast().AddTokenNode(cn.Token())
			}
		}
	}
	return nil
}

func NewNodeFor[T antlr.ParseTree](sf *SourceFile, node T) *NodeBase[T] {
	n := &NodeBase[T]{node: node, sf: sf}

	n.Initialize(n)

	return n
}

type astConversionContext struct {
	parentStack []Node
	result      Node
	sf          *SourceFile
}

func (a *astConversionContext) VisitTerminal(node antlr.TerminalNode) {
	n := NewNodeFor(a.sf, node)
	n.isTerminal = true

	a.addToParent(n)
}

func (a *astConversionContext) VisitErrorNode(node antlr.ErrorNode) {
}

func (a *astConversionContext) EnterEveryRule(ctx antlr.ParserRuleContext) {
	n := NewNodeFor(a.sf, ctx)
	n.isTerminal = false

	if a.sf != nil {
		hidden := a.sf.tokens.GetHiddenTokensToLeft(ctx.GetStart().GetTokenIndex(), 1)

		for _, tk := range hidden {
			if tk.GetChannel() == 1 {
				txt := tk.GetText()
				txt = strings.TrimSpace(txt)

				if txt != "" {
					if strings.HasPrefix(txt, "# TODO:") {
						txt = strings.Replace(txt, "# TODO:", "// TODO:", 1)
					}

					n.comments = append(n.comments, txt)
				}
			}
		}
	}

	a.addToParent(n)

	a.parentStack = append(a.parentStack, n)
}

func (a *astConversionContext) addToParent(n psi.Node) {
	if len(a.parentStack) > 0 {
		parent := a.parentStack[len(a.parentStack)-1]
		n.SetParent(parent)
	}
}

func (a *astConversionContext) ExitEveryRule(ctx antlr.ParserRuleContext) {
	a.result = a.parentStack[len(a.parentStack)-1]
	a.parentStack = a.parentStack[:len(a.parentStack)-1]
}

func AstToPsi(sf *SourceFile, parsed antlr.ParserRuleContext) psi.Node {
	ctx := &astConversionContext{sf: sf}

	walker := antlr.NewParseTreeWalker()
	walker.Walk(ctx, parsed)

	return ctx.result
}
