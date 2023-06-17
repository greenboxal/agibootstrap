package gpt

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"

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

	data, err := json.Marshal(req)

	if err != nil {
		panic(err)
	}

	cctx.SetInput(RequestKey, string(data))
	cctx.SetInput(ObjectiveKey, req.Objective)
	cctx.SetInput(DocumentKey, req.Document)
	cctx.SetInput(ContextKey, req.Context)
	cctx.SetInput(LanguageKey, req.Language)

	return cctx
}

type InvokeOptions struct {
	ForceCodeOutput bool
}

var blockCodeHeaderRegex = regexp.MustCompile("(?m)^```(\\w+)")

// Invoke is a function that invokes the code generator.
func Invoke(ctx context.Context, req Request, opts InvokeOptions) (ast.Node, error) {
	cctx := PrepareContext(ctx, req)

	if err := CodeGeneratorChain.Run(cctx); err != nil {
		return nil, err
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)
	reply := result.Entries[0].Text

	if opts.ForceCodeOutput {
		reply = strings.TrimSpace(reply)
		reply = strings.TrimSuffix(reply, "\n")

		pos := blockCodeHeaderRegex.FindAllString(reply, -1)
		count := len(pos)
		mismatch := count%2 != 0

		if count > 0 && mismatch {
			if strings.HasPrefix(reply, pos[0]) {
				reply = strings.TrimPrefix(reply, pos[0])
			} else if strings.HasSuffix(reply, pos[len(pos)-1]) {
				reply = strings.TrimSuffix(reply, pos[len(pos)-1])
			}
		}

		reply = fmt.Sprintf("```%s\n%s\n```", req.Language, reply)

		lines := strings.Split(reply, "\n")
		for i, line := range lines {
			lines[i] = "\t" + line
		}

		reply = strings.Join(lines, "\n")

	}

	parsedReply := mdutils.ParseMarkdown([]byte(reply))

	return parsedReply, nil
}
