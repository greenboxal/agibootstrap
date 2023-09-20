package gensparse

import (
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

func Pipe[TInToken Token, TNode Node](
	dst ParserStream[sparsing.INodeToken[TNode], TNode],
	src ParserStream[TInToken, TNode],
) {
	fn := ParserNodeHandlerFunc[TInToken, TNode](func(ctx StreamingParserContext[TInToken, TNode], node TNode) error {
		tk := &sparsing.NodeToken[TNode]{Node: node, Path: ctx.CurrentPath()}

		return dst.WriteToken(tk)
	})

	src.PushNodeConsumer(fn)
}

func Identity[TIn, TOut any](node TIn) TOut {
	return any(node).(TOut)
}

func PipeMap[TInToken Token, TIn, TOut Node](
	dst ParserStream[sparsing.INodeToken[TOut], TIn],
	mapper func(TIn) TOut,
) ParserNodeHandler[TInToken, TIn] {
	return ParserNodeHandlerFunc[TInToken, TIn](func(ctx StreamingParserContext[TInToken, TIn], node TIn) error {
		mapped := mapper(node)
		tk := &sparsing.NodeToken[TOut]{Node: mapped, Path: ctx.CurrentPath()}

		return dst.WriteToken(tk)
	})
}
