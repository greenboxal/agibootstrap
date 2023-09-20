package sparsing

type TokenKind int

const (
	TokenKindInvalid TokenKind = 0
	TokenKindEOS               = -2
	TokenKindEOF               = -1
)

type IToken interface {
	GetKind() TokenKind
	GetValue() string
	GetStart() Position
	GetEnd() Position
	GetIndex() int
	GetText() string
}

type INodeToken[T Node] interface {
	IToken

	GetNode() T
	GetPath() []Node
}

type Token struct {
	Kind  TokenKind
	Value string
	Start Position
	End   Position
	Index int

	UserData any
}

func (t *Token) GetKind() TokenKind { return t.Kind }
func (t *Token) GetValue() string   { return t.Value }
func (t *Token) GetStart() Position { return t.Start }
func (t *Token) GetEnd() Position   { return t.End }
func (t *Token) GetIndex() int      { return t.Index }
func (t *Token) GetText() string    { return t.Value }

type NodeToken[T Node] struct {
	Token

	Node T
	Path []Node
}

func (t *NodeToken[T]) GetNode() T      { return t.Node }
func (t *NodeToken[T]) GetPath() []Node { return t.Path }
