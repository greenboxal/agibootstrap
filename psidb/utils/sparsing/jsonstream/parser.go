package jsonstream

import (
	"strconv"

	"github.com/go-errors/errors"

	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

type ParserHandler func(path ParserPath, node Node) error

type Parser struct {
	sparsing.ParserStream

	path ParserPath
}

func NewParser(handler ParserHandler) *Parser {
	p := &Parser{}

	p.ParserStream = sparsing.NewParserStream()
	p.PushLexerHandler(&NodeBase{})
	p.PushTokenParser(&NodeBase{})
	p.PushNodeConsumer(sparsing.ParserNodeHandlerFunc(func(ctx sparsing.StreamingParserContext, node sparsing.Node) error {
		return handler(p.path, node)
	}))

	return p
}

func (p *NodeBase) ConsumeTokenStream(ctx sparsing.StreamingParserContext) error {
	for ctx.RemainingTokens() > 0 {
		switch ctx.CurrentNode().(type) {
		case nil:
			if err := p.acceptValue(ctx); err != nil {
				return err
			}

		case *Object:
			if err := p.acceptObject(ctx); err != nil {
				return err
			}

		case *Pair:
			if err := p.acceptPair(ctx); err != nil {
				return err
			}

		case *Array:
			if err := p.acceptArray(ctx); err != nil {
				return err
			}

		default:
			return errors.New("invalid node type")
		}
	}

	return nil
}

func (p *NodeBase) acceptValue(ctx sparsing.StreamingParserContext) error {
	tk := ctx.NextToken()

	switch tk.GetKind() {
	case TokenKindEOS:
		return nil

	case TokenKindEOF:
		return nil

	case TokenKindOpenObject:
		ctx.BeginComposite(&Object{}, tk)

	case TokenKindOpenArray:
		ctx.BeginComposite(&Array{}, tk)

	case TokenKindString:
		ctx.PushTerminal(&String{}, tk)

	case TokenKindNumber:
		ctx.PushTerminal(&Number{}, tk)

	case TokenKindIdent:
		switch tk.GetValue() {
		case "true", "false":
			ctx.PushTerminal(&Boolean{}, tk)
		case "null":
			ctx.PushTerminal(&Null{}, tk)
		default:
			return ErrInvalidTokenError
		}

	default:
		return ErrInvalidTokenError
	}

	return nil
}

func (p *Parser) Validate() error {
	return nil
}

func (p *NodeBase) acceptString(ctx sparsing.StreamingParserContext) error {
	tk := ctx.ConsumeNextToken(TokenKindString, TokenKindEOS)

	if tk.GetKind() == TokenKindEOS {
		return nil
	}

	ctx.PushTerminal(&String{}, tk)

	return nil
}

func (p *NodeBase) acceptPair(ctx sparsing.StreamingParserContext) error {
	pair := ctx.CurrentNode().(*Pair)

	if pair.Key == nil {
		if ctx.ArgumentStackLength() == 0 {
			if err := p.acceptString(ctx); err != nil {
				return err
			}
		}

		if !ctx.TryPopArgumentTo(&pair.Key) {
			return nil
		}
	}

	if pair.Colon == nil {
		tk := ctx.ConsumeNextToken(TokenKindColon, TokenKindEOS)

		if tk.GetKind() == TokenKindEOS {
			return nil
		}

		pair.Colon = tk
	}

	if pair.Value == nil {
		p.pushPath(pair.Key.GetValue().(string))

		if ctx.ArgumentStackLength() == 0 {
			if err := p.acceptValue(ctx); err != nil {
				return err
			}
		}

		if !ctx.TryPopArgumentTo(&pair.Value) {
			return nil
		}

		p.popPath()
	}

	pair.Start = pair.Key.GetStartToken()
	ctx.EndComposite(pair.Value.GetEndToken())

	return nil
}

func (p *NodeBase) acceptObject(ctx sparsing.StreamingParserContext) error {
	obj := ctx.CurrentNode().(*Object)

	for {
		for ctx.ArgumentStackLength() > 0 {
			n := ctx.PopArgument()

			switch n.(type) {
			case *Pair:
				obj.Pairs = append(obj.Pairs, n.(*Pair))
			default:
				panic(errors.New("invalid node type"))
			}
		}

		tk := ctx.PeekToken(0)

		switch tk.GetKind() {
		case TokenKindEOS:
			return nil

		case TokenKindEOF:
			return errors.New("unexpected EOF")

		case TokenKindCloseObject:
			ctx.EndComposite(ctx.NextToken())
			return nil

		case TokenKindComma:
			if ctx.PeekToken(1).GetKind() == TokenKindCloseObject {
				return errors.New("unexpected comma")
			} else {
				ctx.NextToken()
			}

		default:
			ctx.BeginComposite(&Pair{}, tk)

			if err := p.acceptPair(ctx); err != nil {
				return err
			}
		}
	}
}

func (p *NodeBase) acceptArray(ctx sparsing.StreamingParserContext) error {
	arr := ctx.CurrentNode().(*Array)

	for {
		for ctx.ArgumentStackLength() > 0 {
			n := ctx.PopArgument()

			switch n.(type) {
			case Value:
				arr.Values = append(arr.Values, n.(Value))
			default:
				panic(errors.New("invalid node type"))
			}
		}

		tk := ctx.PeekToken(0)

		switch tk.GetKind() {
		case TokenKindEOS:
			return nil

		case TokenKindCloseArray:
			ctx.EndComposite(ctx.NextToken())
			return nil

		case TokenKindComma:
			if ctx.PeekToken(1).GetKind() == TokenKindCloseArray {
				return errors.New("unexpected comma")
			} else {
				ctx.NextToken()
			}

		default:
			p.pushPath(strconv.Itoa(len(arr.Values)))

			if err := p.acceptValue(ctx); err != nil {
				return err
			}
		}
	}
}

func (p *NodeBase) pushPath(s string) {
	p.path = p.path.Push(s)
}

func (p *NodeBase) popPath() {
	p.path = p.path.Parent()
}
