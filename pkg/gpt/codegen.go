package gpt

import (
	"context"
	"fmt"
	"strings"

	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/hashicorp/go-multierror"
	openai2 "github.com/sashabaranov/go-openai"

	"github.com/greenboxal/agibootstrap/pkg/mdutils"
)

type CodeGeneratorResponse struct {
	MessageLog chat.Message
	CodeBlocks []mdutils.CodeBlock
}

type CodeGenerator struct {
	client *openai.Client
	model  *openai.ChatLanguageModel

	planChain     chain.Chain
	generateChain chain.Chain
	verifyChain   chain.Chain
}

func NewCodeGenerator() *CodeGenerator {
	cg := &CodeGenerator{
		client: oai,
		model:  model,
	}

	cg.planChain = chain.New(
		chain.WithName("CodeGeneratorPLanner"),

		chain.Sequential(
			chat.Predict(
				model,
				CodeGeneratorPlannerPrompt,
			),
		),
	)

	cg.generateChain = chain.New(
		chain.WithName("CodeGenerator"),

		chain.Sequential(
			chat.Predict(
				model,
				CodeGeneratorPrompt,
				chat.WithMaxTokens(4000),

				chat.WithFunction(llm.FunctionDeclaration{
					Name:        "getCodeContext",
					Description: "Returns the code context for the given request, including type definitions for type/method references.",
					Parameters: &llm.FunctionParams{
						Type:     openai2.JSONSchemaTypeObject,
						Required: []string{"code"},
						Properties: map[string]*openai2.JSONSchemaDefine{
							"code": {
								Type: openai2.JSONSchemaTypeString,
							},
						},
					},
				}),

				chat.WithFunction(llm.FunctionDeclaration{
					Name:        "explainCode",
					Description: "Explains the meaning of the given code.",
					Parameters: &llm.FunctionParams{
						Type:     openai2.JSONSchemaTypeObject,
						Required: []string{"code"},
						Properties: map[string]*openai2.JSONSchemaDefine{
							"code": {
								Type: openai2.JSONSchemaTypeString,
							},
						},
					},
				}),

				chat.WithFunction(llm.FunctionDeclaration{
					Name:        "recurseCodeGenerator",
					Description: "Recurse the code generator state machine, calling GPT again with a sub-context.",
					Parameters: &llm.FunctionParams{
						Type:     openai2.JSONSchemaTypeObject,
						Required: []string{"code"},
						Properties: map[string]*openai2.JSONSchemaDefine{
							"code": {
								Type: openai2.JSONSchemaTypeString,
							},
						},
					},
				}),
			),
		),
	)

	cg.verifyChain = chain.New(
		chain.WithName("CodeGeneratorVerifier"),

		chain.Sequential(
			chat.Predict(
				model,
				CodeGeneratorPrompt,
			),
		),
	)

	return cg
}

func (g *CodeGenerator) Generate(ctx context.Context, req CodeGeneratorRequest) (result CodeGeneratorResponse, err error) {
	state := &CodeGeneratorContext{
		gen: g,
		req: req,
	}

	if err = state.Run(ctx); err != nil {
		return
	}

	result.MessageLog = chat.Merge(state.chatHistory...)
	result.CodeBlocks = state.codeBlocks

	return
}

type CodeGeneratorState int

const (
	CodeGenStateInitial CodeGeneratorState = iota
	CodeGenStatePlan
	CodeGenStateGenerate
	CodeGenStateVerify
	CodeGenStateDone
)

type CodeGeneratorContext struct {
	gen *CodeGenerator
	req CodeGeneratorRequest

	state CodeGeneratorState

	plan        string
	chatHistory []chat.Message
	codeBlocks  []mdutils.CodeBlock

	errors []error
}

func (s *CodeGeneratorContext) Load(ctx chain.ChainContext) (chat.Message, error) {
	return chat.Merge(s.chatHistory...), nil
}

func (s *CodeGeneratorContext) Append(ctx chain.ChainContext, msg chat.Message) error {
	s.chatHistory = append(s.chatHistory, msg)

	return nil
}

func (s *CodeGeneratorContext) Run(ctx context.Context) (err error) {
	defer func() {
		var err error

		for _, e := range s.errors {
			err = multierror.Append(err, e)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		switch s.state {
		case CodeGenStateInitial:
			s.setState(CodeGenStatePlan)
		case CodeGenStatePlan:
			s.stepPlan(ctx)
		case CodeGenStateGenerate:
			s.stepGenerate(ctx)
		case CodeGenStateVerify:
			s.stepVerify(ctx)
		case CodeGenStateDone:
			return
		}
	}
}

func (s *CodeGeneratorContext) stepPlan(ctx context.Context) {
	cctx := PrepareContext(ctx, s.req)

	cctx.SetInput(chat.MemoryContextKey, s)

	if err := s.gen.planChain.Run(cctx); err != nil {
		s.abort(err)
		return
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)

	s.plan = result.Entries[0].Text

	s.setState(CodeGenStateGenerate)
}

func (s *CodeGeneratorContext) stepGenerate(ctx context.Context) {
	cctx := PrepareContext(ctx, s.req)

	cctx.SetInput(PlanKey, s.plan)
	cctx.SetInput(chat.MemoryContextKey, s)

	if err := s.gen.generateChain.Run(cctx); err != nil {
		s.abort(err)
		return
	}

	result := chain.Output(cctx, chat.ChatReplyContextKey)

	s.processResult(result)
}

func (s *CodeGeneratorContext) stepVerify(ctx context.Context) {
	if len(s.codeBlocks) == 0 {
		s.setState(CodeGenStateGenerate)
		return
	}

	s.setState(CodeGenStateDone)
}

func (s *CodeGeneratorContext) processResult(result chat.Message) {
	reply := result.Entries[0].Text
	reply = s.sanitizeCodeBlockReply(reply)
	replyRoot := mdutils.ParseMarkdown([]byte(reply))

	blocks := mdutils.ExtractCodeBlocks(replyRoot)

	s.codeBlocks = append(s.codeBlocks, blocks...)

	s.setState(CodeGenStateVerify)
}

func (s *CodeGeneratorContext) abort(err error) {
	s.reportError(err)
	s.setState(CodeGenStateDone)
}

func (s *CodeGeneratorContext) reportError(err error) {
	s.errors = append(s.errors, err)
}

func (s *CodeGeneratorContext) setState(state CodeGeneratorState) {
	s.state = state
}

func (s *CodeGeneratorContext) sanitizeCodeBlockReply(reply string) string {
	reply = strings.TrimSpace(reply)
	reply = strings.TrimSuffix(reply, "\n")

	pos := blockCodeHeaderRegex.FindAllString(reply, -1)
	count := len(pos)
	mismatch := count%2 != 0

	if count > 0 && mismatch {
		if strings.HasPrefix(reply, pos[0]) {
			reply = strings.TrimPrefix(reply, pos[0])
			reply = fmt.Sprintf("```%s\n%s\n```", s.req.Language, reply)
		} else if strings.HasSuffix(reply, pos[len(pos)-1]) {
			reply = strings.TrimSuffix(reply, pos[len(pos)-1])
			reply = fmt.Sprintf("```%s\n%s\n```", s.req.Language, reply)
		}
	}

	return reply
}
