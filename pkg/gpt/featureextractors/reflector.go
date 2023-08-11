package featureextractors

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
	"github.com/invopop/jsonschema"
	"github.com/jaswdr/faker"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
)

type ReflectOptions struct {
	History []*thoughtdb.Thought
	Query   string

	ExampleInput  any
	ExampleOutput any
}

var globalReflectorMutex = &sync.Mutex{}
var globalReflector = &jsonschema.Reflector{}

func GetSchemaForType[T any]() *jsonschema.Schema {
	typ := reflect.TypeOf((*T)(nil)).Elem()

	return ReflectSchemaForType(typ)
}

func ReflectSchemaForType(typ reflect.Type) *jsonschema.Schema {
	globalReflectorMutex.Lock()
	defer globalReflectorMutex.Unlock()

	return globalReflector.ReflectFromType(typ)
}

type initializerIface interface {
	psi.Node

	Init(self psi.Node)
}

func Reflect[T any](ctx context.Context, req ReflectOptions) (def T, _ chat.Message, _ error) {
	for i := 0; i < 10; i++ {
		res, reply, err := reflectSingle[T](ctx, req)

		if err == nil {
			return res, reply, nil
		}
	}

	return def, chat.Message{}, fmt.Errorf("failed to reflect")
}

func reflectSingle[T any](ctx context.Context, req ReflectOptions) (def T, _ chat.Message, _ error) {
	schema := GetSchemaForType[T]()

	if req.ExampleOutput == nil {
		typ := reflect.TypeOf((*T)(nil)).Elem()

		req.ExampleOutput = reflect.New(typ).Elem().Interface()

		faker.New().Struct().Fill(req.ExampleOutput)
	}

	schemaJson, err := json.Marshal(schema)

	if err != nil {
		return def, chat.Message{}, err
	}

	examplesJson, err := json.Marshal(req.ExampleOutput)

	if err != nil {
		return def, chat.Message{}, err
	}

	historyText := ""

	for _, msg := range req.History {
		historyText += fmt.Sprintf("[%s]:\n%s\n", msg.From.Name, msg.Text)
	}

	msgs := make([]*thoughtdb.Thought, 0, len(req.History)+3)

	msgs = append(msgs, &thoughtdb.Thought{
		From: thoughtdb.CommHandle{
			Role: msn.RoleSystem,
		},

		Text: fmt.Sprintf(
			"You will be tasked to reply to a request about the chat history. "+
				"You should reply with a json object respecting the following schema:\n"+"```jsonschema\n%s\n```\n"+
				"For example:\n"+
				"Given this input: ```json\n%s\n```\n"+
				"Correct output: ```json\n%s\n```\n",
			schemaJson,
			req.ExampleInput,
			examplesJson,
		),
	})

	msgs = append(msgs, req.History...)

	msgs = append(msgs, &thoughtdb.Thought{
		From: thoughtdb.CommHandle{
			Name: "User",
			Role: msn.RoleUser,
		},

		Text: fmt.Sprintf(
			"<<< History >>>\n```\n%s```\n"+
				"<<< Schema >>>\n```jsonschema\n%s```\n"+
				"<<< Request >>>\nPlease answer the following request about the chat history and reply respecting the data format defined in the schema above.\n"+
				"- %s\n",
			historyText,
			schemaJson,
			req.Query,
		),
	})

	msgs = append(msgs, &thoughtdb.Thought{
		From: thoughtdb.CommHandle{
			Name: "Assistant",
			Role: msn.RoleAI,
		},

		Text: "```json\n",
	})

	prompt := &SimplePromptTemplate{
		Messages: msgs,
	}

	cctx := chain.NewChainContext(ctx)

	stepChain := chain.New(
		chain.WithName("FeatureReflector"),

		chain.Sequential(
			chat.Predict(
				gpt.GlobalModel,
				prompt,
				chat.WithMaxTokens(1024),
			),
		),
	)

	if err := stepChain.Run(cctx); err != nil {
		return def, chat.Message{}, err
	}

	reply := chain.Output(cctx, chat.ChatReplyContextKey)
	sanitized := gpt.SanitizeCodeBlockReply(reply.Entries[0].Text, "json")
	replyRoot := mdutils.ParseMarkdown([]byte(sanitized))
	blocks := mdutils.ExtractCodeBlocks(replyRoot)

	var jsonBlock []byte

	if len(blocks) > 0 {
		jsonBlock = []byte(blocks[0].Code)
	} else {
		jsonBlock = []byte(reply.Entries[0].Text)
	}

	if err := json.Unmarshal(jsonBlock, &def); err != nil {
		return def, chat.Message{}, err
	}

	if init, ok := any(def).(initializerIface); ok {
		init.Init(init)
	}

	return def, reply, nil
}
