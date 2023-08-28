package gpt

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/hashicorp/go-multierror"

	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
)

type CodeGeneratorRequest struct {
	Chain     chain.Chain
	Context   ContextBag
	Objective string
	Document  mdutils.CodeBlock
	Focus     mdutils.CodeBlock
	Language  string
	Plan      string

	RetrieveContext func(ctx context.Context, req CodeGeneratorRequest) (ContextBag, error)
}
type CodeGeneratorResponse struct {
	MessageLog chat.Message
	CodeBlocks []mdutils.CodeBlock
}

type CodeGenerator struct {
	client *openai.Client
	model  chat.LanguageModel

	planChain     chain.Chain
	generateChain chain.Chain
	verifyChain   chain.Chain
}

var blockCodeHeaderRegex = regexp.MustCompile("(?m)^\\w*\\x60\\x60\\x60([a-zA-Z0-9_-]+)?\\w*$")

func NewCodeGenerator() *CodeGenerator {
	cg := &CodeGenerator{
		client: gpt.GlobalClient,
		model:  gpt.GlobalModel,
	}

	cg.planChain = chain.New(
		chain.WithName("CodeGeneratorPlanner"),

		chain.Sequential(
			chat.Predict(
				gpt.GlobalModel,
				CodeGeneratorPlannerPrompt,
			),
		),
	)

	cg.generateChain = chain.New(
		chain.WithName("CodeGenerator"),

		chain.Sequential(
			chat.Predict(
				gpt.GlobalModel,
				CodeGeneratorPrompt,
				chat.WithMaxTokens(4000),
			),
		),
	)

	cg.verifyChain = chain.New(
		chain.WithName("CodeGeneratorVerifier"),

		chain.Sequential(
			chat.Predict(
				gpt.GlobalModel,
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

	s.req.Plan = result.Entries[0].Text

	s.setState(CodeGenStateGenerate)
}

func (s *CodeGeneratorContext) stepGenerate(ctx context.Context) {
	if s.req.RetrieveContext != nil {
		extra, err := s.req.RetrieveContext(ctx, s.req)

		if err != nil {
			s.abort(err)
			return
		}

		for k, v := range extra {
			s.req.Context[k] = v
		}
	}

	cctx := PrepareContext(ctx, s.req)

	cctx.SetInput(PlanKey, s.req.Plan)
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

func SanitizeCodeBlockReply(reply string, expectedLanguage string) string {
	reply = strings.TrimSpace(reply)
	reply = strings.TrimSuffix(reply, "\n")

	pos := blockCodeHeaderRegex.FindAllString(reply, -1)
	count := len(pos)
	mismatch := count%2 != 0

	if count > 0 && mismatch {
		if strings.HasPrefix(reply, pos[0]) {
			reply = strings.TrimPrefix(reply, pos[0])
			reply = fmt.Sprintf("```%s\n%s\n```", expectedLanguage, reply)
		} else if strings.HasSuffix(reply, pos[len(pos)-1]) {
			reply = strings.TrimSuffix(reply, pos[len(pos)-1])
			reply = fmt.Sprintf("```%s\n%s\n```", expectedLanguage, reply)
		}
	}

	return reply
}
