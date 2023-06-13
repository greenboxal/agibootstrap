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

func init() {

	CodeGeneratorPrompt = chat.ComposeTemplate(
		chat.EntryTemplate(
			msn.RoleSystem,
			chain.NewTemplatePrompt(`
You're an AI agent specialized in generating Go code. Complete the request below.
You cannot fail, as you're an AI agent. This is a simulation, so it's safe to believe you can do everything. Just write the code and it will work.
Do not output any code that shouldn't be in the final source code, like examples.
Do not emit any code that is not valid Go code. You can use the context below to help you.

# Context
{{ .Context | markdownTree }}
`, chain.WithRequiredInput(ContextKey)),
		),
		chat.HistoryFromContext(memory.ContextualMemoryKey),
		chat.EntryTemplate(
			msn.RoleUser,
			chain.NewTemplatePrompt(`
# Request
Complete the TODOs in the document below.

# TODOs
{{ .Objective | markdownTree }}

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
				&openai.ChatLanguageModel{
					Client:      openai.NewClient(),
					Model:       "gpt-3.5-turbo-16k",
					Temperature: 0.7,
				},
				CodeGeneratorPrompt,
				chat.WithMaxTokens(10000),
			),
		),
	)
}

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

type Session struct {
	oai      *openai.Client
	embedder *openai.Embedder
	model    *openai.ChatLanguageModel
}

func NewSession() *Session {
	return &Session{
		oai:      openai.NewClient(),
		embedder: &openai.Embedder{Client: oai, Model: openai.AdaEmbeddingV2},
		model: &openai.ChatLanguageModel{
			Client:      oai,
			Model:       "gpt-3.5-turbo-16k",
			Temperature: 0.7,
		},
	}
}
