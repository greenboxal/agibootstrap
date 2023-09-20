package autoform

import (
	"context"
	"encoding/json"

	"github.com/go-errors/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/sashabaranov/go-openai"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/text/mdutils"
	"github.com/greenboxal/agibootstrap/psidb/services/chat"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
	jsonstream2 "github.com/greenboxal/agibootstrap/psidb/utils/sparsing/jsonstream"
)

type ResultParser interface {
	ParseChoice(ctx context.Context, choice PromptResponseChoice) error
}

type StreamedResultParser interface {
	ResultParser

	ParseChoiceStreamed(ctx context.Context, choice openai.ChatCompletionStreamChoice) error
	Error() error
}

type ResultParserFunc func(ctx context.Context, choice PromptResponseChoice) error

func (r ResultParserFunc) ParseChoice(ctx context.Context, choice PromptResponseChoice) error {
	return r(ctx, choice)
}

type RecoverableError struct {
	Err error

	ErrorPosition       sparsing.Position
	RecoverablePosition sparsing.Position
}

func (r RecoverableError) Error() string { return r.Err.Error() }
func (r RecoverableError) Unwrap() error { return r.Err }

func ParseToLog(log ChatLog, baseMessage *chat.Message) ResultParser {
	return ResultParserFunc(func(ctx context.Context, choice PromptResponseChoice) error {
		return log.AcceptChoice(ctx, baseMessage, choice)
	})
}

type multiParser struct {
	parsers []ResultParser
}

func (m multiParser) Error() error {
	var merr error

	for _, parser := range m.parsers {
		parser, ok := parser.(StreamedResultParser)

		if !ok {
			continue
		}

		if err := parser.Error(); err != nil {
			merr = multierror.Append(merr, err)
		}
	}

	return merr
}

func (m multiParser) ParseChoice(ctx context.Context, choice PromptResponseChoice) error {
	for _, parser := range m.parsers {
		if err := parser.ParseChoice(ctx, choice); err != nil {
			return err
		}
	}

	return nil
}

func (m multiParser) ParseChoiceStreamed(ctx context.Context, choice openai.ChatCompletionStreamChoice) error {
	for _, parser := range m.parsers {
		parser, ok := parser.(StreamedResultParser)

		if !ok {
			continue
		}

		if err := parser.ParseChoiceStreamed(ctx, choice); err != nil {
			return err
		}
	}

	return nil
}

func MultiParser(parsers ...ResultParser) ResultParser {
	return &multiParser{parsers: parsers}
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

type PromptParserTarget string

const (
	PromptParserTargetText         PromptParserTarget = "text"
	PromptParserTargetFunctionCall PromptParserTarget = "function_call"
)

func ParseJson(target PromptParserTarget, result any) ResultParser {
	return &streamingJsonParser{target: target, results: []any{result}}
}

type streamingJsonParser struct {
	target  PromptParserTarget
	results []any
	err     error

	parser *jsonstream2.Parser
}

func (s *streamingJsonParser) Error() error {
	return s.err
}

func (s *streamingJsonParser) ParseChoiceStreamed(ctx context.Context, choice openai.ChatCompletionStreamChoice) error {
	if s.parser == nil {
		s.parser = jsonstream2.NewParser(func(path jsonstream2.ParserPath, node jsonstream2.Node) error {
			//fmt.Printf("%s: %s\n", path, repr.String(node))
			return nil
		})
	}

	txt := ""

	if s.target == PromptParserTargetFunctionCall {
		if choice.Delta.FunctionCall == nil {
			return nil
		}

		txt = choice.Delta.FunctionCall.Arguments
	} else if s.target == PromptParserTargetText {
		txt = choice.Delta.Content
	}

	if txt == "" {
		return nil
	}

	previousPosition := s.parser.Position()

	if _, err := s.parser.Write([]byte(txt)); err != nil {
		var perr *sparsing.ParsingError

		if errors.As(err, &perr) {
			err = &RecoverableError{
				Err:                 err,
				ErrorPosition:       perr.Position,
				RecoverablePosition: previousPosition,
			}
		}

		s.err = multierror.Append(s.err, err)

		return err
	}

	return nil
}

func (s *streamingJsonParser) FinishParseChoiceStreamed(ctx context.Context, choice PromptResponseChoice) error {
	var node jsonstream2.Value

	if s.err != nil {
		return s.err
	}

	if s.parser == nil {
		return nil
	}

	if err := s.parser.Close(); err != nil {
		return err
	}

	if !s.parser.TryPopArgumentTo(&node) {
		return nil
	}

	result := node.GetValue()

	data, err := json.Marshal(result)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, s.results[choice.Index]); err != nil {
		return err
	}

	return nil
}

func (s *streamingJsonParser) ParseChoice(ctx context.Context, choice PromptResponseChoice) error {
	var jsonBlock []byte

	if s.parser != nil {
		if err := s.FinishParseChoiceStreamed(ctx, choice); err != nil {
			return err
		}
	}

	if s.target == PromptParserTargetFunctionCall {
		if choice.Message.FunctionCall == nil {
			return nil
		}

		jsonBlock = []byte(choice.Message.FunctionCall.Arguments)
	} else {
		sanitized := gpt.SanitizeCodeBlockReply(choice.Message.Text, "json")
		replyRoot := mdutils.ParseMarkdown([]byte(sanitized))
		blocks := mdutils.ExtractCodeBlocks(replyRoot)

		if len(blocks) > 0 {
			jsonBlock = []byte(blocks[0].Code)
		} else {
			jsonBlock = []byte(choice.Message.Text)
		}
	}

	if err := json.Unmarshal(jsonBlock, s.results[choice.Index]); err != nil {
		return err
	}

	return nil
}

func ParseJsonStreamed(target PromptParserTarget, result ...any) StreamedResultParser {
	return &streamingJsonParser{target: target, results: result}
}

type StreamingParser[T sparsing.Node] struct {
	target  PromptParserTarget
	results []*T
	err     error

	marker string
	parser sparsing.ParserStream
}

func NewStreamingParser[T sparsing.Node](target PromptParserTarget, parser sparsing.ParserStream, marker string, result ...*T) *StreamingParser[T] {
	return &StreamingParser[T]{parser: parser, marker: marker, target: target, results: result}
}

func (s *StreamingParser[T]) Error() error {
	return s.err
}

func (s *StreamingParser[T]) ParseChoiceStreamed(ctx context.Context, choice openai.ChatCompletionStreamChoice) error {
	txt := ""

	if s.target == PromptParserTargetFunctionCall {
		if choice.Delta.FunctionCall == nil {
			return nil
		}

		txt = choice.Delta.FunctionCall.Arguments
	} else if s.target == PromptParserTargetText {
		txt = choice.Delta.Content
	}

	if txt == "" {
		return nil
	}

	if _, err := s.parser.Write([]byte(txt)); err != nil {
		var perr *sparsing.ParsingError

		if errors.As(err, &perr) {
			s.parser.Recover(perr.RecoverablePosition.Offset)

			err = &RecoverableError{
				Err:                 err,
				ErrorPosition:       perr.Position,
				RecoverablePosition: s.parser.Position(),
			}
		}

		s.err = err

		return err
	}

	return nil
}

func (s *StreamingParser[T]) FinishParseChoiceStreamed(ctx context.Context, choice PromptResponseChoice) error {
	if s.err != nil {
		return s.err
	}

	if s.parser == nil {
		return nil
	}

	if err := s.parser.Close(); err != nil {
		var perr *sparsing.ParsingError

		if errors.As(err, &perr) {
			err = &RecoverableError{
				Err:           err,
				ErrorPosition: perr.Position,
			}
		}

		return err
	}

	return nil
}

func (s *StreamingParser[T]) ParseChoice(ctx context.Context, choice PromptResponseChoice) error {
	var inputBlock []byte

	if s.parser != nil {
		if err := s.FinishParseChoiceStreamed(ctx, choice); err != nil {
			return err
		}
	}

	if s.target == PromptParserTargetFunctionCall {
		if choice.Message.FunctionCall == nil {
			return nil
		}

		inputBlock = []byte(choice.Message.FunctionCall.Arguments)
	} else {
		sanitized := gpt.SanitizeCodeBlockReply(choice.Message.Text, s.marker)
		replyRoot := mdutils.ParseMarkdown([]byte(sanitized))
		blocks := mdutils.ExtractCodeBlocks(replyRoot)

		if len(blocks) > 0 {
			inputBlock = []byte(blocks[0].Code)
		} else {
			inputBlock = []byte(choice.Message.Text)
		}
	}

	if _, err := s.parser.Write(inputBlock); err != nil {
		return err
	}

	if err := s.FinishParseChoiceStreamed(ctx, choice); err != nil {
		return err
	}

	return nil
}
