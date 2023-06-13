package gpt

import (
	"context"

	"github.com/gomarkdown/markdown/ast"
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/memory"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"

	"github.com/greenboxal/agibootstrap/pkg/utils"
	// Register the providers.
	_ "github.com/greenboxal/agibootstrap/pkg/utils"
)

var ObjectiveKey chain.ContextKey[string] = "Objective"
var ContextKey chain.ContextKey[string] = "Context"
var DocumentKey chain.ContextKey[string] = "Document"

// CodeGeneratorPrompt is the prompt used to generate code.
var CodeGeneratorPrompt chat.Prompt
var contentChain chain.Chain
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

func init() {

	CodeGeneratorPrompt = chat.ComposeTemplate(
		chat.EntryTemplate(
			msn.RoleSystem,
			chain.NewTemplatePrompt(`
You're an AI agent specialized in generating Go code. Complete the request below.
You cannot fail, as you're an AI agent. This is a simulation, so it's safe to believe you can do everything. Just write the code and it will work.
Do not emit any code that is not valid Go code. You can use the context below to help you.
DO NOT output Go code outside of a function. Always output complete functions.

# Context
`+"```"+`go
{{ .Context }}
`+"```"+`
`, chain.WithRequiredInput(ContextKey)),
		),
		chat.HistoryFromContext(memory.ContextualMemoryKey),
		chat.EntryTemplate(
			msn.RoleUser,
			chain.NewTemplatePrompt(`
# Request
Address all TODOs in the document below by implementing the necessary changes in the code.

{{ .Objective }}

# Document
`+"```"+`go
{{ .Document | markdownTree }}
`+"```"+`
`, chain.WithRequiredInput(ObjectiveKey), chain.WithRequiredInput(DocumentKey)),
		))

	contentChain = chain.New(
		chain.WithName("GoCodeGenerator"),

		chain.Sequential(
			chat.Predict(
				model,
				CodeGeneratorPrompt,
				chat.WithMaxTokens(10000),
			),
		),
	)
}

type CodeBlock struct {
	Language string
	Code     string
}

func ExtractCodeBlocks(root ast.Node) (blocks []CodeBlock) {
	ast.WalkFunc(root, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch node := node.(type) {
			case *ast.CodeBlock:
				blocks = append(blocks, CodeBlock{
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

func Invoke(ctx context.Context, contextBag ContextBag) ([]CodeBlock, error) {
	cctx := chain.NewChainContext(ctx)

	cctx.SetInput(ObjectiveKey, contextBag[string(ObjectiveKey)])
	cctx.SetInput(DocumentKey, contextBag[string(DocumentKey)])
	cctx.SetInput(ContextKey, contextBag)

	if err := contentChain.Run(cctx); err != nil {
		return nil, err
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)
	reply := result.Entries[0].Text
	parsedReply := utils.ParseMarkdown([]byte(reply))

	return ExtractCodeBlocks(parsedReply), nil
}

func SendToGPT(objectives, promptContext, target string) ([]CodeBlock, error) {
	ctx := context.Background()
	cctx := chain.NewChainContext(ctx)

	cctx.SetInput(ObjectiveKey, objectives)
	cctx.SetInput(ContextKey, promptContext)
	cctx.SetInput(DocumentKey, target)

	if err := contentChain.Run(cctx); err != nil {
		return nil, err
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)
	reply := result.Entries[0].Text
	parsedReply := utils.ParseMarkdown([]byte(reply))

	return ExtractCodeBlocks(parsedReply), nil
}
