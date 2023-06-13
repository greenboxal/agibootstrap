package gpt

import (
	"context"
	"html"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/md"
	"github.com/gomarkdown/markdown/parser"
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/memory"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
)

var ObjectiveKey chain.ContextKey[string] = "Objective"
var ContextKey chain.ContextKey[string] = "Context"
var DocumentKey chain.ContextKey[string] = "Document"

// CodeGeneratorPrompt is the prompt used to generate code.
// TODO: Make this the best prompt so an LLM like GPT-3.5-TURBO (you) can generate code.
var CodeGeneratorPrompt = chat.ComposeTemplate(
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

var oai = openai.NewClient()
var embedder = &openai.Embedder{
	Client: oai,
	Model:  openai.AdaEmbeddingV2,
}
var model = &openai.ChatLanguageModel{
	Client:      oai,
	Model:       "gpt-3.5-turbo",
	Temperature: 0.7,
}

var contentChain = chain.New(
	chain.WithName("GoCodeGenerator"),

	chain.Sequential(
		chat.Predict(
			&openai.ChatLanguageModel{
				Client:      openai.NewClient(),
				Model:       "gpt-3.5-turbo",
				Temperature: 0.7,
			},
			CodeGeneratorPrompt,
		),
	),
)

func SendToGPT(objectives, promptContext, target string) (string, error) {
	ctx := context.Background()
	cctx := chain.NewChainContext(ctx)

	cctx.SetInput(ObjectiveKey, objectives)
	cctx.SetInput(ContextKey, promptContext)
	cctx.SetInput(DocumentKey, target)

	if err := contentChain.Run(cctx); err != nil {
		return "", err
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)
	reply := result.Entries[0].Text
	codeOutput := ""
	parsedReply := ParseMarkdown([]byte(reply))

	ast.WalkFunc(parsedReply, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			switch node := node.(type) {
			case *ast.CodeBlock:
				codeOutput += string(node.Literal)
			}
		}

		return ast.GoToNext
	})

	// Check if codeOutput is possibly HTML escaped and unescape it
	codeOutput = html.UnescapeString(codeOutput)

	return codeOutput, nil
}

func FormatMarkdown(node ast.Node) []byte {
	return markdown.Render(node, md.NewRenderer())
}

func ParseMarkdown(md []byte) ast.Node {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	return p.Parse(md)
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
			Model:       "gpt-3.5-turbo",
			Temperature: 0.7,
		},
	}
}
