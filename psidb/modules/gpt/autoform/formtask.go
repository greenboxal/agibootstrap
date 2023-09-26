package autoform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alecthomas/repr"
	"github.com/go-errors/errors"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/hashicorp/go-multierror"
	"github.com/invopop/jsonschema"
	openai2 "github.com/sashabaranov/go-openai"
	"github.com/xeipuuv/gojsonschema"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type FormTaskToolHandler[Ctx, T any] func(ctx context.Context, form Ctx, choice PromptResponseChoice, arg T) error

type IFormTaskTool[Ctx any] interface {
	PromptBuilderTool

	Execute(ctx context.Context, form Ctx, choice PromptResponseChoice) error
}

type FormTaskTool[Ctx, T any] struct {
	Name        string
	Description string
	Handler     FormTaskToolHandler[Ctx, T]
}

func (f *FormTaskTool[Ctx, T]) ToolName() string {
	return f.Name
}

func (f *FormTaskTool[Ctx, T]) ToolDefinition() *openai.FunctionDefinition {
	var args llm.FunctionParams

	args.Type = "object"

	requestType := typesystem.GetType[T]()
	def := requestType.JsonSchema()
	def = typesystem.FlattenJsonSchema(typesystem.Universe(), def)

	if def.Type == "object" {
		args.Type = def.Type
		args.Required = def.Required
		args.Properties = map[string]*jsonschema.Schema{}

		for _, key := range def.Properties.Keys() {
			v, _ := def.Properties.Get(key)

			args.Properties[key] = v.(*jsonschema.Schema)
		}
	} else {
		args.Type = "object"
		args.Properties = map[string]*jsonschema.Schema{
			"request": def,
		}
		args.Required = []string{"request"}
	}

	return &openai.FunctionDefinition{
		Name:        f.Name,
		Description: f.Description,
		Parameters:  &args,
	}
}

func (f *FormTaskTool[Ctx, T]) GetName() string { return f.Name }

func (f *FormTaskTool[Ctx, T]) Execute(ctx context.Context, form Ctx, choice PromptResponseChoice) error {
	var args T

	callForm := NewForm(WithSchemaFor[T]())
	raw := []byte(choice.Message.FunctionCall.Arguments)

	if err := callForm.UnmarshalJSON(raw); err != nil {
		return err
	}

	err := callForm.UnmarshalTo(&args)

	if err != nil {
		return err
	}

	return f.Handler(ctx, form, choice, args)
}

type FormTask struct {
	DirectedTask

	form *Form

	pastHistory PromptMessageSource
	formHistory []openai.ChatCompletionMessage

	lastChoice *PromptResponseChoice

	epoch  int
	errors []gojsonschema.ResultError

	tools map[string]IFormTaskTool[*FormTask]
}

func MakeTool[Ctx, T any](name string, description string, handler FormTaskToolHandler[Ctx, T]) *FormTaskTool[Ctx, T] {
	return &FormTaskTool[Ctx, T]{Name: name, Description: description, Handler: handler}
}

type SetFieldValue struct {
	JsonPointerString string           `json:"jsonPointerString"`
	Value             *json.RawMessage `json:"value"`
}

func NewFormTask(form *Form, history PromptMessageSource) *FormTask {
	ft := &FormTask{
		form:        form,
		pastHistory: history,
		tools:       map[string]IFormTaskTool[*FormTask]{},
	}

	ft.DirectedTask = *NewDirectedTask(ft)

	ft.tools["SetFieldValue"] = MakeTool(
		"SetFieldValue",
		"Sets the value of a field",
		func(ctx context.Context, form *FormTask, choice PromptResponseChoice, arg SetFieldValue) error {
			var value any

			if arg.JsonPointerString == "" {
				return errors.New("invalid json pointer")
			}

			if arg.Value == nil || len(*arg.Value) == 0 {
				return errors.New("invalid value")
			}

			if err := json.Unmarshal(*arg.Value, &value); err != nil {
				return err
			}

			if err := form.form.SetField(arg.JsonPointerString, value); err != nil {
				return err
			}

			return nil

			/*result, err := form.form.Validate()

			if err != nil {
				return err
			}

			if !result.Valid() {
				var merr error

				for _, err := range result.Errors() {
					merr = multierror.Append(merr, errors.Wrap(err, 1))
				}

				return merr
			}

			return nil*/
		},
	)

	ft.tools["MergeFieldValues"] = MakeTool(
		"MergeFieldValues",
		"Deeply merges the values of a field",
		func(ctx context.Context, form *FormTask, choice PromptResponseChoice, arg SetFieldValue) error {
			other := NewForm()
			subForm, err := form.form.SubForm(arg.JsonPointerString)

			if err != nil {
				return err
			}

			if err := other.UnmarshalJSON(*arg.Value); err != nil {
				return err
			}

			if err := subForm.MergeFields(other.Value); err != nil {
				return err
			}

			result, err := subForm.Validate()

			if err != nil {
				return err
			}

			if !result.Valid() {
				var merr error

				for _, err := range result.Errors() {
					merr = multierror.Append(merr, errors.Wrap(err, 1))
				}

				return merr
			}

			return nil
		},
	)

	ft.tools["ClearFieldValue"] = MakeTool(
		"MergeFieldValues",
		"Deeply merges the values of a field",
		func(ctx context.Context, form *FormTask, choice PromptResponseChoice, arg SetFieldValue) error {
			return form.form.SetField(arg.JsonPointerString, nil)
		},
	)

	ft.tools["ClearFields"] = MakeTool(
		"ClearFields",
		"Clears the form",
		func(ctx context.Context, form *FormTask, choice PromptResponseChoice, arg SetFieldValue) error {
			form.form.Clear()

			return nil
		},
	)

	return ft
}

func (f *FormTask) createPromptBuilder(ctx context.Context) *PromptBuilder {
	pb := NewPromptBuilder()
	pb.WithClient(gpt.GlobalClient)
	pb.WithModelOptions(gpt.ModelOptions{
		Model:       gpt.String("gpt-3.5-turbo"),
		Temperature: gpt.Float32(0.7),
		MaxTokens:   gpt.Int(1024),
	})

	pb.AppendMessageSources(PromptBuilderHookHistory, f.pastHistory)

	for _, msg := range f.formHistory {
		pb.AppendModelMessage(PromptBuilderHookPostHistory, msg)
	}

	for _, tool := range f.tools {
		pb.WithTools(tool)
	}

	return pb
}

func (f *FormTask) OnPrepare(ctx context.Context) error {

	return nil
}

func (f *FormTask) OnStep(ctx context.Context) error {
	var lastMsg *openai.ChatCompletionMessage

	repr.Println(f.form.Value)

	fmt.Printf("EPOCH: %d\tSTEP: %d\t", f.epoch, f.step)

	if len(f.formHistory) > 0 {
		lastMsg = &f.formHistory[len(f.formHistory)-1]
	}

	if f.lastChoice != nil {
		if lastMsg != nil && lastMsg.Role == "function" {
			fmt.Printf("Function Continuation: %s\n", lastMsg.Name)
			return f.continueFunction(ctx)
		} else if f.lastChoice.Reason == openai.FinishReasonFunctionCall {
			fmt.Printf("Function Call: %s\n", f.lastChoice.Message.FunctionCall.Name)
			return f.unroll(ctx, f.lastChoice.Message.FunctionCall)
		}
	}

	fmt.Printf("Nudge\n")
	return f.nudge(ctx)
}

func (f *FormTask) unroll(ctx context.Context, call *chat.FunctionCall) error {
	tool := f.tools[call.Name]

	if tool == nil {
		return errors.New("unknown tool")
	}

	err := tool.Execute(ctx, f, *f.lastChoice)

	if err != nil {
		return err
	}

	f.formHistory = append(f.formHistory, openai.ChatCompletionMessage{
		Role:    "function",
		Name:    call.Name,
		Content: `Success! Next step...`,
	})

	return nil
}

func (f *FormTask) continueFunction(ctx context.Context) error {
	return f.predictNext(ctx, f.createPromptBuilder(ctx))
}

func (f *FormTask) nudge(ctx context.Context) error {
	f.epoch++

	pb := f.createPromptBuilder(ctx)

	formJson, err := f.form.MarshalJSON()

	if err != nil {
		return err
	}

	if len(f.errors) > 0 {
		fmt.Printf("Errors: %s\n", f.errors)

		errorsJson, err := json.Marshal(f.errors)

		if err != nil {
			return err
		}

		pb.AppendModelMessage(PromptBuilderHookFocus, openai.ChatCompletionMessage{
			Role: "user",
			Content: `Form: ` + string(formJson) + `

Current Errors:` + string(errorsJson) + `

Fix the errors above in the form.
`,
		})
	} else {
		pb.AppendModelMessage(PromptBuilderHookFocus, openai.ChatCompletionMessage{
			Role: "user",
			Content: `Form: ` + string(formJson) + `

Complete the form above following the declared schema. Use the tools available to help you complete the form.
`,
		})
	}

	return f.predictNext(ctx, pb)
}

func (f *FormTask) predictNext(ctx context.Context, pb *PromptBuilder) error {
	schemaJson, err := typesystem.Universe().CompileBundleFor(f.form.Schema).MarshalJSON()

	if err != nil {
		return err
	}

	pb.AppendModelMessage(PromptBuilderHookPreFocus, openai.ChatCompletionMessage{
		Role: "system",
		Content: `
You are an assistant completing a form. The form is incomplete. Complete the form below following the declared schema. Use the tools available to help you complete the form.

Tools:
* ` + "`" + `SetFieldValue({ jsonPointerString: string }, value: any)` + "`" + `: sets the value of a field
* ` + "`" + `MergeFieldValues({ jsonPointerString: string }, value: any)` + "`" + `: deeply merges the values of a field
* ` + "`" + `ClearFieldValue(({ jsonPointerString: string }, value: any)` + "`" + `: clears the value of a field
* ` + "`" + `ClearFields()` + "`" + `: clears the form

Schema: ` + string(schemaJson),
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

			f.acceptChoice(choice)

			return nil
		}),
	)

	return pb.ExecuteAndParse(ctx, parser)
}

func (f *FormTask) HandleError(err error) error {
	msg := openai.ChatCompletionMessage{
		Role:    "user",
		Content: "Operation failed: Error: " + err.Error(),
	}

	if f.lastChoice != nil && f.lastChoice.Reason == openai.FinishReasonFunctionCall {
		msg.Role = "function"
		msg.Name = f.lastChoice.Message.FunctionCall.Name
	}

	f.formHistory = append(f.formHistory, msg)

	fmt.Printf("ERROR: %s\n", err.Error())

	return nil
}

func (f *FormTask) CheckComplete(ctx context.Context) (bool, error) {
	result, err := f.form.Validate()

	if err != nil {
		return false, err
	}

	f.errors = result.Errors()

	if !result.Valid() {
		return false, nil
	}

	return true, nil
}

func (f *FormTask) acceptChoice(choice PromptResponseChoice) {
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

	f.formHistory = append(f.formHistory, msg)
	f.lastChoice = &choice
}

type FunctionCallParser struct {
	ParsedFunctionName      string
	ParsedFunctionArguments string

	SelectedTool PromptBuilderTool
	Choice       *PromptResponseChoice
}

func (f *FunctionCallParser) ParseChoice(ctx context.Context, choice PromptResponseChoice) error {
	if choice.Reason != openai.FinishReasonFunctionCall {
		return nil
	}

	if choice.Message.FunctionCall == nil {
		return nil
	}

	f.ParsedFunctionName = choice.Message.FunctionCall.Name
	f.ParsedFunctionArguments = choice.Message.FunctionCall.Arguments
	f.Choice = &choice

	return nil
}

func OnChoiceParsed(fn func(ctx context.Context, choice PromptResponseChoice) error) ResultParser {
	return ResultParserFunc(fn)
}

type ChunkSniffer func(ctx context.Context, choice openai2.ChatCompletionStreamChoice) error

func (c ChunkSniffer) Error() error { return nil }
func (c ChunkSniffer) ParseChoiceStreamed(ctx context.Context, choice openai2.ChatCompletionStreamChoice) error {
	return c(ctx, choice)
}
func (c ChunkSniffer) ParseChoice(ctx context.Context, choice PromptResponseChoice) error {
	return nil
}
