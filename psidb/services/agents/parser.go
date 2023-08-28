package agents

import (
	"context"
	"encoding/json"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
	"github.com/greenboxal/agibootstrap/pkg/typesystem"
)

type ResultParser interface {
	ParseChoice(ctx context.Context, choice PromptResponseChoice) error
}

type ResultParserFunc func(ctx context.Context, choice PromptResponseChoice) error

func ParseToLog(log ChatLog, baseMessage *Message) ResultParser {
	return ResultParserFunc(func(ctx context.Context, choice PromptResponseChoice) error {
		return log.AcceptChoice(ctx, baseMessage, choice)
	})
}

func MultiParser(parsers ...ResultParser) ResultParser {
	return ResultParserFunc(func(ctx context.Context, choice PromptResponseChoice) error {
		for _, parser := range parsers {
			if err := parser.ParseChoice(ctx, choice); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r ResultParserFunc) ParseChoice(ctx context.Context, choice PromptResponseChoice) error {
	return r(ctx, choice)
}

func ParseString(str *string) ResultParser {
	return ResultParserFunc(func(ctx context.Context, choice PromptResponseChoice) error {
		*str = choice.Message.Text

		return nil
	})
}

func ParseJsonWithPrototype[T any](result *T) ResultParser {
	typ := typesystem.GetType[T]()

	return ResultParserFunc(func(ctx context.Context, choice PromptResponseChoice) error {
		sanitized := gpt.SanitizeCodeBlockReply(choice.Message.Text, "json")
		replyRoot := mdutils.ParseMarkdown([]byte(sanitized))
		blocks := mdutils.ExtractCodeBlocks(replyRoot)

		var jsonBlock []byte

		if len(blocks) > 0 {
			jsonBlock = []byte(blocks[0].Code)
		} else {
			jsonBlock = []byte(choice.Message.Text)
		}

		node, err := ipld.DecodeUsingPrototype(jsonBlock, dagjson.Decode, typ.IpldPrototype())

		if err != nil {
			return err
		}

		unwrapped, ok := typesystem.TryUnwrap[T](node)

		if !ok {
			return errors.New("type mismatch")
		}

		*result = unwrapped

		return nil
	})
}

func ParseJson(result any) ResultParser {
	return ResultParserFunc(func(ctx context.Context, choice PromptResponseChoice) error {
		sanitized := gpt.SanitizeCodeBlockReply(choice.Message.Text, "json")
		replyRoot := mdutils.ParseMarkdown([]byte(sanitized))
		blocks := mdutils.ExtractCodeBlocks(replyRoot)

		var jsonBlock []byte

		if len(blocks) > 0 {
			jsonBlock = []byte(blocks[0].Code)
		} else {
			jsonBlock = []byte(choice.Message.Text)
		}

		if err := json.Unmarshal(jsonBlock, result); err != nil {
			return err
		}

		return nil
	})
}
