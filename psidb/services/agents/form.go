package agents

import (
	"encoding/json"
	"fmt"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/invopop/jsonschema"
	"github.com/xeipuuv/gojsonschema"

	"github.com/greenboxal/agibootstrap/psidb/services/chat"
)

type Form struct {
	Schema *jsonschema.Schema
	Data   map[string]interface{}
	Errors []gojsonschema.ResultError

	validationSchema *gojsonschema.Schema
	hasError         bool
	isComplete       bool
}

func NewForm(schema *jsonschema.Schema) *Form {
	return &Form{
		Schema: schema,
	}
}

func (f *Form) HasError() bool {
	return f.hasError
}

func (f *Form) IsComplete() bool {
	return f.isComplete
}

func (f *Form) Invalidate() {
	f.isComplete = false
}

func (f *Form) getValidationSchema() *gojsonschema.Schema {
	if f.validationSchema != nil {
		return f.validationSchema
	}

	validationSchema, err := gojsonschema.NewSchema(gojsonschema.NewGoLoader(f.Schema))

	if err != nil {
		panic(err)
	}

	f.validationSchema = validationSchema

	return f.validationSchema
}

func (f *Form) Validate() (bool, error) {
	if f.isComplete {
		return true, nil
	}

	result, err := f.getValidationSchema().Validate(gojsonschema.NewGoLoader(&f.Data))

	if err != nil {
		return false, err
	}

	if !result.Valid() {
		f.Errors = result.Errors()

		return false, nil
	}

	f.Errors = nil
	f.isComplete = true

	return true, nil
}

func (f *Form) Fix(ctx *ThreadContext) (bool, error) {
	if ok, err := f.Validate(); err != nil || ok {
		return ok, err
	}

	for _, err := range f.Errors {
		if _, err := f.FixError(ctx, err); err != nil {
			return false, err
		}
	}

	return f.Validate()
}

func (f *Form) FixError(ctx *ThreadContext, formError gojsonschema.ResultError) (bool, error) {
	var patch any

	currentValue, err := json.Marshal(f.Data)

	if err != nil {
		return false, err
	}

	schema, err := f.Schema.MarshalJSON()

	if err != nil {
		return false, err
	}

	schemaMsg := chat.NewMessage(chat.MessageKindEmit)
	schemaMsg.From.Role = msn.RoleSystem
	schemaMsg.Text = fmt.Sprintf(`"Response JSONSchema:
`+"```jsonschema"+`
%s
`+"```"+`
`, string(schema))

	queryMsg := chat.NewMessage(chat.MessageKindEmit)
	queryMsg.From.Role = msn.RoleUser
	queryMsg.Text = fmt.Sprintf(`"
**Current Data:**
`+"```json"+`
%s
`+"```"+`

**Error:** %s

Fix the error above in the form.
`, string(currentValue), formError.String())

	pb := ctx.BuildPrompt()
	pb.AppendMessageSources(PromptBuilderHookHistory, MessageSourceFromChatHistory(ctx.History))
	pb.AppendMessage(PromptBuilderHookPreFocus, schemaMsg)
	pb.SetFocus(queryMsg)

	err = pb.ExecuteAndParse(ctx.Ctx, MultiParser(
		ParseToLog(ctx.Log, ctx.BaseMessage),
		ParseJson(PromptParserTargetText, &patch),
	))

	if err != nil {
		return false, err
	}

	if err := f.PatchField(ctx, formError.Context(), patch); err != nil {
		return false, err
	}

	return true, nil
}

func (f *Form) FillFields(ctx *ThreadContext, fields ...*gojsonschema.JsonContext) error {
	schema, err := f.Schema.MarshalJSON()

	if err != nil {
		return err
	}

	schemaMsg := chat.NewMessage(chat.MessageKindEmit)
	schemaMsg.From.Role = msn.RoleSystem
	schemaMsg.Text = fmt.Sprintf(`"Response JSONSchema:
`+"```jsonschema"+`
%s
`+"```"+`
`, string(schema))

	for _, field := range fields {
		var patch any

		currentValue, err := json.Marshal(f.Data)

		if err != nil {
			return err
		}

		queryMsg := chat.NewMessage(chat.MessageKindEmit)
		queryMsg.From.Role = msn.RoleUser
		queryMsg.Text = fmt.Sprintf(`"
**Current Data:**
`+"```json"+`
%s
`+"```"+`

Fill the following field in the form: %s
`, string(currentValue), field.String())

		pb := ctx.BuildPrompt()
		pb.AppendMessageSources(PromptBuilderHookHistory, MessageSourceFromChatHistory(ctx.History))
		pb.AppendMessage(PromptBuilderHookPreFocus, schemaMsg)
		pb.SetFocus(queryMsg)

		err = pb.ExecuteAndParse(ctx.Ctx, MultiParser(
			ParseToLog(ctx.Log, ctx.BaseMessage),
			ParseJson(PromptParserTargetText, &patch),
		))

		if err != nil {
			return err
		}

		if err := f.PatchField(ctx, field, patch); err != nil {
			return err
		}
	}

	return nil
}

func (f *Form) FillAll(ctx *ThreadContext) error {
	currentValue, err := json.Marshal(f.Data)

	if err != nil {
		return err
	}

	schema, err := f.Schema.MarshalJSON()

	if err != nil {
		return err
	}

	queryMsg := chat.NewMessage(chat.MessageKindEmit)
	queryMsg.From.Role = msn.RoleUser
	queryMsg.Text = fmt.Sprintf(`"Response JSONSchema:
`+"```jsonschema"+`
%s
`+"```"+`
Complete the below following the JSONSchema above.

Current Data:
`+"```json"+`
%s
`+"```"+`

Complete the form below following the JSONSchema above.
`, string(schema), string(currentValue))

	pb := ctx.BuildPrompt()
	pb.AppendMessageSources(PromptBuilderHookHistory, MessageSourceFromChatHistory(ctx.History))
	pb.SetFocus(queryMsg)

	f.Invalidate()

	err = pb.ExecuteAndParse(ctx.Ctx, MultiParser(
		ParseToLog(ctx.Log, ctx.BaseMessage),
		ParseJson(PromptParserTargetText, &f.Data),
	))

	if err != nil {
		return err
	}

	return nil
}

func (f *Form) PatchField(ctx *ThreadContext, context *gojsonschema.JsonContext, patch any) error {
	return f.Marshal(patch)
}

func (f *Form) ParseJson(data []byte) error {
	return json.Unmarshal(data, &f.Data)
}

func (f *Form) Marshal(result any) error {
	data, err := json.Marshal(result)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, &f.Data)
}

func (f *Form) Unmarshal(result any) error {
	data, err := json.Marshal(f.Data)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, result)
}

func (f *Form) ToJSON() ([]byte, error) {
	return json.Marshal(f.Data)
}
