package autoform

import (
	"context"
	"fmt"
	"strconv"

	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"
	"github.com/pkg/errors"
	openai2 "github.com/sashabaranov/go-openai"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
)

type TextComprehension struct {
	DirectedTask

	Content        string
	TokensPerChunk int
	TokensPerPage  int

	Log          []openai.ChatCompletionMessage
	Observations []string

	CurrentIndex int

	tools     map[string]IFormTaskTool[*TextComprehension]
	tokenizer tokenizers.BasicTokenizer
	tokens    tokenizers.Tokens
}

type AppendObservation struct {
	Observations []string `json:"observation"`
}

type NextPage struct {
	Observations []string `json:"observation"`
}

func NewTextComprehension(
	tokenizer tokenizers.BasicTokenizer,
	content string,
	linesPerChunk, linesPerPage int,
) *TextComprehension {
	fc := &TextComprehension{
		tokenizer: tokenizer,
		tools:     map[string]IFormTaskTool[*TextComprehension]{},

		Content:        content,
		TokensPerChunk: linesPerChunk,
		TokensPerPage:  linesPerPage,
	}

	fc.DirectedTask = *NewDirectedTask(fc)

	fc.tools["AppendObservation"] = MakeTool(
		"AppendObservation",
		"Append one or more observations to the log",

		func(ctx context.Context, form *TextComprehension, choice PromptResponseChoice, arg AppendObservation) error {
			fc.Observations = append(fc.Observations, arg.Observations...)
			fc.CurrentIndex += fc.TokensPerChunk

			return nil
		},
	)

	fc.tools["NextPage"] = MakeTool(
		"NextPage",
		"Go to the next page",

		func(ctx context.Context, form *TextComprehension, choice PromptResponseChoice, arg NextPage) error {
			fc.Observations = append(fc.Observations, arg.Observations...)
			fc.CurrentIndex += fc.TokensPerPage

			return nil
		},
	)

	tokens, err := fc.tokenizer.GetTokens([]byte(fc.Content))

	if err != nil {
		panic(err)
	}

	fc.tokens = tokens

	return fc
}

func (t *TextComprehension) OnPrepare(ctx context.Context) error {

	return nil
}

func (t *TextComprehension) OnStep(ctx context.Context) error {
	firstLine := t.CurrentIndex
	lastLine := t.CurrentIndex + t.TokensPerChunk

	if lastLine > t.tokens.Len() {
		lastLine = t.tokens.Len()
	}

	tokens := t.tokens.Slice(firstLine, lastLine)
	content := tokens.String()

	pb := t.createPromptBuilder(ctx)

	pb.AppendModelMessage(PromptBuilderHookPreFocus, openai.ChatCompletionMessage{
		Role: "system",
		Content: `You are reading through a large document are trying to find the answer to a question. You are currently on page ` + strconv.Itoa(t.CurrentIndex) + `.
Append observations about all facts you have learned so far to the log above. You can also use the "Next Page" tool to go to the next page.
`,
	})

	pb.AppendModelMessage(PromptBuilderHookFocus, openai.ChatCompletionMessage{
		Role:    "user",
		Content: content,
	})

	fnCall := &FunctionCallParser{}

	parser := MultiParser(
		fnCall,

		ChunkSniffer(func(ctx context.Context, choice openai2.ChatCompletionStreamChoice) error {
			fmt.Printf("%s", choice.Delta.Content)

			if choice.Delta.FunctionCall != nil {
				fmt.Printf("%s", choice.Delta.FunctionCall.Arguments)
			}

			return nil
		}),

		OnChoiceParsed(func(ctx context.Context, choice PromptResponseChoice) error {
			fmt.Printf("\n")

			if choice.Reason == openai.FinishReasonFunctionCall {
				fmt.Printf("Function Call: %s\n", choice.Message.FunctionCall.Name)
			}

			t.acceptChoice(choice)

			return nil
		}),
	)

	if err := pb.ExecuteAndParse(ctx, parser); err != nil {
		return err
	}

	if fnCall.ParsedFunctionName != "" {
		tool := t.tools[fnCall.ParsedFunctionName]

		if tool == nil {
			return errors.New("unknown tool")
		}

		err := tool.Execute(ctx, t, *fnCall.Choice)

		if err != nil {
			return err
		}

		t.Log = append(t.Log, openai.ChatCompletionMessage{
			Role:    "function",
			Name:    fnCall.ParsedFunctionName,
			Content: `Success! Next step...`,
		})
	}

	return nil
}

func (t *TextComprehension) createPromptBuilder(ctx context.Context) *PromptBuilder {
	pb := NewPromptBuilder()
	pb.WithClient(gpt.GlobalClient)
	pb.WithModelOptions(gpt.ModelOptions{
		Model:       gpt.String("gpt-3.5-turbo-16k"),
		Temperature: gpt.Float32(0.7),
		MaxTokens:   gpt.Int(1024),
	})

	for _, tool := range t.tools {
		pb.WithTools(tool)
	}

	logs := t.Log

	if len(logs) > 2 {
		logs = logs[len(logs)-2:]
	}

	for _, msg := range logs {
		pb.AppendModelMessage(PromptBuilderHookHistory, msg)
	}

	return pb
}

func (t *TextComprehension) CheckComplete(ctx context.Context) (bool, error) {
	return t.CurrentIndex >= t.tokens.Len(), nil
}

func (t *TextComprehension) HandleError(err error) error {
	return nil
}
func (t *TextComprehension) acceptChoice(choice PromptResponseChoice) {
	msg := openai.ChatCompletionMessage{
		Role:    "assistant",
		Content: choice.Message.Text,
	}

	if choice.Message.FunctionCall != nil {
		msg.FunctionCall = &openai.FunctionCall{
			Name:      choice.Message.FunctionCall.Name,
			Arguments: choice.Message.FunctionCall.Arguments,
		}
	}

	t.Log = append(t.Log, msg)
}
