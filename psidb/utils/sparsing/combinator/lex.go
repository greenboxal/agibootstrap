package combinator

import (
	"regexp"

	"github.com/go-errors/errors"
	"golang.org/x/exp/slices"

	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

type LexerNode interface {
	BasicNode

	PeekMatch(ctx sparsing.StreamingLexerContext) bool
	Match(ctx sparsing.StreamingLexerContext) error

	CanAccept(ctx sparsing.StreamingParserContext) bool
	Accept(ctx sparsing.StreamingParserContext) error
}

type LexerNodeBuilder interface {
	LexerNode

	Build() LexerNode
}

type TerminalParser struct {
	kind  sparsing.TokenKind
	regex *regexp.Regexp
}

func (t *TerminalParser) CanAccept(ctx sparsing.StreamingParserContext) bool {
	return ctx.PeekToken(0).GetKind() == t.kind
}

func (t *TerminalParser) Accept(ctx sparsing.StreamingParserContext) error {
	ctx.PushTerminal(nil, ctx.ConsumeNextToken(t.kind))
	return nil
}

func (t *TerminalParser) IsTerminal() bool {
	return true
}

func NewTerminal(kind sparsing.TokenKind, regex string) *TerminalParser {
	return &TerminalParser{
		kind:  kind,
		regex: regexp.MustCompile(regex),
	}
}

func (tp *TerminalParser) PeekMatch(ctx sparsing.StreamingLexerContext) bool {
	_, ok := ctx.PeekMatchRegexp(tp.regex)

	return ok
}

func (tp *TerminalParser) Match(ctx sparsing.StreamingLexerContext) error {
	m, ok := ctx.TryMatchRegexp(tp.regex)

	if !ok {
		return errors.New("failed to match")
	}

	ctx.PushSingle(tp.kind, m[0])

	return nil
}

type SequenceParser struct {
	nodes []LexerNode
}

func (sp *SequenceParser) IsTerminal() bool {
	for _, node := range sp.nodes {
		if !node.IsTerminal() {
			return false
		}
	}

	return true
}

func NewSequenceParser(nodes ...LexerNode) *SequenceParser {
	return &SequenceParser{
		nodes: nodes,
	}
}

func (sp *SequenceParser) Append(node ...LexerNode) {
	sp.nodes = append(sp.nodes, node...)
}

func (sp *SequenceParser) Build() LexerNode {
	return &SequenceParser{
		nodes: slices.Clone(sp.nodes),
	}
}

func (sp *SequenceParser) PeekMatch(ctx sparsing.StreamingLexerContext) bool {
	//FIXME: Doesn't work when peeking
	for _, node := range sp.nodes {
		if !node.PeekMatch(ctx) {
			return false
		}
	}

	return true
}

func (sp *SequenceParser) Match(ctx sparsing.StreamingLexerContext) error {
	for _, node := range sp.nodes {
		if err := node.Match(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (sp *SequenceParser) CanAccept(ctx sparsing.StreamingParserContext) bool {
	//FIXME: Doesn't work when peeking
	for _, node := range sp.nodes {
		if !node.CanAccept(ctx) {
			return false
		}
	}

	return true
}

func (sp *SequenceParser) Accept(ctx sparsing.StreamingParserContext) error {
	for _, node := range sp.nodes {
		if err := node.Accept(ctx); err != nil {
			return err
		}
	}

	return nil
}

type ChoiceParser struct {
	nodes []LexerNode
}

func (cp *ChoiceParser) IsTerminal() bool {
	for _, node := range cp.nodes {
		if !node.IsTerminal() {
			return false
		}
	}

	return true
}

func NewChoiceParser(nodes ...LexerNode) *ChoiceParser {
	return &ChoiceParser{
		nodes: nodes,
	}
}

func (cp *ChoiceParser) AddChoice(node LexerNode) {
	cp.nodes = append(cp.nodes, node)
}

func (cp *ChoiceParser) Build() LexerNode {
	return &ChoiceParser{
		nodes: slices.Clone(cp.nodes),
	}
}

func (cp *ChoiceParser) PeekMatch(ctx sparsing.StreamingLexerContext) bool {
	for _, node := range cp.nodes {
		if node.PeekMatch(ctx) {
			return true
		}
	}

	return false
}

func (cp *ChoiceParser) Match(ctx sparsing.StreamingLexerContext) error {
	for _, node := range cp.nodes {
		if node.PeekMatch(ctx) {
			return node.Match(ctx)
		}
	}

	return errors.New("no node can accept")
}

func (cp *ChoiceParser) CanAccept(ctx sparsing.StreamingParserContext) bool {
	for _, node := range cp.nodes {
		if node.CanAccept(ctx) {
			return true
		}
	}

	return false
}

func (cp *ChoiceParser) Accept(ctx sparsing.StreamingParserContext) error {
	for _, node := range cp.nodes {
		if node.CanAccept(ctx) {
			return node.Accept(ctx)
		}
	}

	return errors.New("no node can accept")
}

func NewOptionalParser(node LexerNode) LexerNode {
	return &OptionalParser{
		node: node,
	}
}

func (op *OptionalParser) PeekMatch(ctx sparsing.StreamingLexerContext) bool {
	return op.node.PeekMatch(ctx)
}

func (op *OptionalParser) Match(ctx sparsing.StreamingLexerContext) error {
	if op.node.PeekMatch(ctx) {
		return op.node.Match(ctx)
	}

	return nil
}

type OptionalParser struct {
	node LexerNode
}

func (op *OptionalParser) IsTerminal() bool {
	return op.node.IsTerminal()
}

func (op *OptionalParser) CanAccept(ctx sparsing.StreamingParserContext) bool {
	return op.node.CanAccept(ctx)
}

func (op *OptionalParser) Accept(ctx sparsing.StreamingParserContext) error {
	if op.node.CanAccept(ctx) {
		return op.node.Accept(ctx)
	}

	return nil
}

type RepeatParser struct {
	node     LexerNode
	min, max int
}

func (rp *RepeatParser) IsTerminal() bool {
	return rp.node.IsTerminal()
}

func NewRepeatParser(node LexerNode, min, max int) *RepeatParser {
	return &RepeatParser{
		node: node,
		min:  min,
		max:  max,
	}
}

func (rp *RepeatParser) Build() LexerNode {
	return rp
}

func (rp *RepeatParser) PeekMatch(ctx sparsing.StreamingLexerContext) bool {
	//FIXME: Doesn't work when peeking in order to account for min
	return rp.node.PeekMatch(ctx)
}

func (rp *RepeatParser) Match(ctx sparsing.StreamingLexerContext) error {
	count := 0

	for {
		if count >= rp.max {
			break
		}

		if count >= rp.min {
			if !rp.node.PeekMatch(ctx) {
				break
			}
		}

		if err := rp.node.Match(ctx); err != nil {
			return err
		}

		count++
	}

	if rp.min <= 0 {
		if !rp.node.PeekMatch(ctx) {
			return nil
		}
	}

	return nil
}

func (rp *RepeatParser) CanAccept(ctx sparsing.StreamingParserContext) bool {
	//FIXME: Doesn't work when peeking in order to account for min
	return rp.node.CanAccept(ctx)
}

func (rp *RepeatParser) Accept(ctx sparsing.StreamingParserContext) error {
	count := 0

	for {
		if count >= rp.max {
			break
		}

		if count >= rp.min {
			if !rp.node.CanAccept(ctx) {
				break
			}
		}

		if err := rp.node.Accept(ctx); err != nil {
			return err
		}

		count++
	}

	if rp.min <= 0 {
		if !rp.node.CanAccept(ctx) {
			return nil
		}
	}

	return nil
}
