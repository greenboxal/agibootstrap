package pylang

import (
	"strings"

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

	sf         *SourceFile
	start, end antlr.Token
}

func (nb *NodeBase[T]) IsContainer() bool            { return len(nb.Children()) > 0 }
func (nb *NodeBase[T]) IsLeaf() bool                 { return len(nb.Children()) == 0 }
func (nb *NodeBase[T]) Ast() antlr.ParserRuleContext { return nb.node }
func (nb *NodeBase[T]) Comments() []string           { return nb.comments }

func (nb *NodeBase[T]) Initialize(self Node) {
	nb.NodeBase.Init(self, "")
}

func (nb *NodeBase[T]) Update() {
	if nb.IsValid() {
		return
	}

	nb.NodeBase.Update()

	for i := 0; i < nb.node.GetChildCount(); i++ {
		nb.node.RemoveLastChild()
	}

	for _, c := range nb.Children() {
		nb.node.AddChild(c.(Node).Ast())
	}
}

func NewNodeFor(sf *SourceFile, node antlr.ParserRuleContext) *NodeBase[antlr.ParserRuleContext] {
	n := &NodeBase[antlr.ParserRuleContext]{node: node, sf: sf}

	n.Initialize(n)

	return n
}

type astConversionContext struct {
	parentStack []Node
	result      Node
	sf          *SourceFile
}

func (a *astConversionContext) VisitTerminal(node antlr.TerminalNode) {
}

func (a *astConversionContext) VisitErrorNode(node antlr.ErrorNode) {
}

func (a *astConversionContext) EnterEveryRule(ctx antlr.ParserRuleContext) {
	n := NewNodeFor(a.sf, ctx)

	if a.sf != nil {
		hidden := a.sf.tokens.GetHiddenTokensToLeft(ctx.GetStart().GetTokenIndex(), 2)

		for _, tk := range hidden {
			if tk.GetChannel() == 2 {
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
	ctx := &astConversionContext{sf: sf}

	walker := antlr.NewParseTreeWalker()
	walker.Walk(ctx, parsed)

	return ctx.result
}
