package gpt

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/chain"

	// Register the providers.
	_ "github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

// PrepareContext prepares the context for the given request.
//
// Parameters:
// - ctx: The context.Context for the operation.
// - req: The CodeGeneratorRequest containing the input data.
//
// Returns:
// - chain.ChainContext: The prepared generateChain context.
func PrepareContext(ctx context.Context, req CodeGeneratorRequest) chain.ChainContext {
	cctx := chain.NewChainContext(ctx)

	cctx.SetInput(ObjectiveKey, req.Objective)
	cctx.SetInput(DocumentKey, req.Document)
	cctx.SetInput(FocusKey, req.Focus)
	cctx.SetInput(ContextKey, req.Context)
	cctx.SetInput(LanguageKey, req.Language)

	return cctx
}
