package gpt

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/mdutils"
	// Register the providers.
	_ "github.com/greenboxal/agibootstrap/pkg/mdutils"
)

var oai = openai.NewClient()
var embedder = &openai.Embedder{
	Client: oai,
	Model:  openai.AdaEmbeddingV2,
}
var model = &openai.ChatLanguageModel{
	Client:      oai,
	Model:       "gpt-3.5-turbo-16k",
	Temperature: 0.7,
}

type ContextBag map[string]any

type Request struct {
	Chain     chain.Chain
	Context   ContextBag
	Objective string
	Document  mdutils.CodeBlock
	Language  string
}

// PrepareContext prepares the context for the given request.
//
// Parameters:
// - ctx: The context.Context for the operation.
// - req: The Request containing the input data.
//
// Returns:
// - chain.ChainContext: The prepared chain context.
func PrepareContext(ctx context.Context, req Request) chain.ChainContext {
	cctx := chain.NewChainContext(ctx)

	cctx.SetInput(ObjectiveKey, req.Objective)
	cctx.SetInput(DocumentKey, req.Document)
	cctx.SetInput(ContextKey, req.Context)
	cctx.SetInput(LanguageKey, req.Language)

	return cctx
}

// Invoke is a function that invokes the code generator.
func Invoke(ctx context.Context, req Request) ([]mdutils.CodeBlock, error) {
	cctx := PrepareContext(ctx, req)

	if err := CodeGeneratorChain.Run(cctx); err != nil {
		return nil, err
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)
	reply := result.Entries[0].Text
	parsedReply := mdutils.ParseMarkdown([]byte(reply))

	return mdutils.ExtractCodeBlocks(parsedReply), nil
}
