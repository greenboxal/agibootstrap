package gpt

import (
	"context"

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
`, chain.WithRequiredInput(ContextKey))),

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
`, chain.WithRequiredInput(ObjectiveKey), chain.WithRequiredInput(DocumentKey))),
)

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
			model,
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

	return codeOutput, nil
}

//func Summarize(objective, document string) (string, error) {
//	const summarizerPromptTemplate = `
//Generate a summary of the document:
//
//Objective: {{ .Objective }}
//Document: {{ .Document }}
//`
//
//	ctx := context.Background()
//	cctx := chain.NewChainContext(ctx)
//
//	cctx.SetInput(ObjectiveKey, objective)
//	cctx.SetInput(DocumentKey, document)
//
//	summarizerChain := chain.Sequential(
//		chat.Predict(
//			model,
//			chat.ComposeTemplate(summarizerPromptTemplate, chain.WithRequiredInput(ObjectiveKey), chain.WithRequiredInput(DocumentKey)),
//		),
//	)
//
//	if err := summarizerChain.Run(cctx); err != nil {
//		return "", err
//	}
//
//	result := chain.Output(cctx, chat.ChatReplyContextKey)
//	reply := result.Entries[0].Text
//	return reply, nil
//}

func FormatMarkdown(node ast.Node) []byte {
	return markdown.Render(node, md.NewRenderer())
}

func ParseMarkdown(md []byte) ast.Node {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	return p.Parse(md)
}

type Session struct {
}

func NewSession() *Session {
	// T2ODO: Implement a Session type that can be used to store the context of a conversation.
	// T2ODO: It should include (and replace) the globals `oai`, `embedder`, and `model` that are defined above.

	return nil
}
