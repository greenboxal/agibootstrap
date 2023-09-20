package sparsing

func Pipe(p ParserStream) ParserNodeHandler {
	return ParserNodeHandlerFunc(func(ctx StreamingParserContext, node Node) error {
		return p.WriteToken(&NodeToken[Node]{Node: node, Path: ctx.CurrentPath()})
	})
}
