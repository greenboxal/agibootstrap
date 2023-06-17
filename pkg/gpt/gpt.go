package gpt

import (
	"context"

	"github.com/gomarkdown/markdown/ast"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/utils"

	// Register the providers.
	_ "github.com/greenboxal/agibootstrap/pkg/utils"
)

var oai = openai.NewClient()
var embedder = &openai.Embedder{
	Client: oai,
	Model:  openai.AdaEmbeddingV2,
}
var model = &openai.ChatLanguageModel{
	Client:      oai,
	Model:       "gpt-3.5-turbo-16k",
	Temperature: 1,
}

// ExtractCodeBlocks traverses the given AST and extracts all code blocks.
// It returns a slice of CodeBlock objects, each representing a code block
// with its language and code content.
func ExtractCodeBlocks(root ast.Node) (blocks []psi.CodeBlock) {
	ast.WalkFunc(root, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch node := node.(type) {
			case *ast.CodeBlock:
				blocks = append(blocks, psi.CodeBlock{
					Language: string(node.Info),
					Code:     string(node.Literal),
				})
			}
		}

		return ast.GoToNext
	})

	return
}

type ContextBag map[string]any

type Request struct {
	Chain     chain.Chain
	Context   ContextBag
	Objective string
	Document  string
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

	return cctx
}

// Invoke is a function that invokes the code generator.
func Invoke(ctx context.Context, req Request) ([]psi.CodeBlock, error) {
	cctx := PrepareContext(ctx, req)

	if err := CodeGeneratorChain.Run(cctx); err != nil {
		return nil, err
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)
	reply := result.Entries[0].Text
	parsedReply := utils.ParseMarkdown([]byte(reply))

	return ExtractCodeBlocks(parsedReply), nil
}

// SendToGPT sends objectives, prompt context, and target to GPT for code generation.
// It returns a slice of CodeBlock objects and an error.
func SendToGPT(objectives, promptContext, target string) ([]psi.CodeBlock, error) {
	ctx := context.Background()
	cctx := chain.NewChainContext(ctx)

	cctx.SetInput(ObjectiveKey, objectives)
	cctx.SetInput(ContextKey, promptContext)
	cctx.SetInput(DocumentKey, target)

	if err := CodeGeneratorChain.Run(cctx); err != nil {
		return nil, err
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)
	reply := result.Entries[0].Text
	parsedReply := utils.ParseMarkdown([]byte(reply))

	return ExtractCodeBlocks(parsedReply), nil
}
