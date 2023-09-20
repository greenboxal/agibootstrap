package sparsing

type Node interface {
	GetStartToken() IToken
	GetEndToken() IToken
}

type LexeableNode interface {
	Node
	StreamingLexerHandler
}

type ParseableNode interface {
	Node
	ParserTokenHandler
}

type TerminalNode interface {
	Node

	SetTerminalToken(tk IToken) error
}

type CompositeNode interface {
	Node

	SetStartToken(tk IToken)
	SetEndToken(tk IToken)
}

type NodeBase struct {
	Start IToken `json:"start"`
	End   IToken `json:"end"`
}

func (n *NodeBase) GetStartToken() IToken { return n.Start }
func (n *NodeBase) GetEndToken() IToken   { return n.End }

type TerminalNodeBase struct {
	NodeBase
}

func (n *TerminalNodeBase) GetText() string { return n.Start.GetText() }
func (n *TerminalNodeBase) SetTerminalToken(tk IToken) error {
	n.Start = tk
	n.End = tk

	return nil
}

type CompositeNodeBase struct {
	NodeBase
}

func (n *CompositeNodeBase) SetStartToken(tk IToken) { n.Start = tk }
func (n *CompositeNodeBase) SetEndToken(tk IToken)   { n.End = tk }
