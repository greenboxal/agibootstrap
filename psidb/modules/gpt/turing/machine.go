package turing

import (
	"context"

	"github.com/greenboxal/agibootstrap/psidb/modules/gpt/autoform"
)

type Machine struct {
	stack Stack[*Frame]
}

func NewMachine() *Machine {
	return &Machine{}
}

func (m *Machine) CurrentFrame() *Frame {
	return m.stack.Peek()
}

func (m *Machine) PushFrame(f *Frame) {
	m.stack.Push(f)
}

func (m *Machine) PopFrame() *Frame {
	return m.stack.Pop()
}

func (m *Machine) Eval(ctx context.Context, op MicroOperation) error {
	frame := m.CurrentFrame()

	switch op.Op {
	case MicroOpInfer:
		return m.processInfer(ctx, frame)

	case MicroOpPushIn:
		frame.InStack.Push(op.V)

	case MicroOpPopIn:
		frame.InStack.Pop()

	case MicroOpPopIntToOut:
		frame.OutStack.Push(frame.InStack.Pop())

	case MicroOpPushOut:
		frame.OutStack.Push(op.V)

	case MicroOpPopOut:
		frame.OutStack.Pop()

	case MicroOpCall:
		m.PushFrame(NewFrame(frame))

	case MicroOpReturn:
		next := m.stack.LA(1)

		if next != nil {
			next.InStack.Append(frame.OutStack.All()...)
		}

		m.PopFrame()

	case MicroOpAbort:
		m.PopFrame()
	}

	return nil
}

func (m *Machine) processInfer(ctx context.Context, frame *Frame) error {
	parser := autoform.MultiParser(
		autoform.OnChoiceParsed(func(ctx context.Context, choice autoform.PromptResponseChoice) error {
			if choice.Reason == autoform.FinishReasonFunctionCall {
				frame.OutStack.Push(Value{
					Kind: ValueKindReference,
				})
			} else {

			}

			return nil
		}),
	)

	pb := autoform.NewPromptBuilder()
	err := pb.ExecuteAndParse(ctx, parser)

	if err != nil {
		return err
	}

	return nil
}
