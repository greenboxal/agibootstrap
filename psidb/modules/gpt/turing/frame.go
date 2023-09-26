package turing

type Frame struct {
	Parent *Frame

	InStack  ValueStack
	OutStack ValueStack
}

func NewFrame(parent *Frame) *Frame {
	f := &Frame{
		Parent: parent,
	}

	if parent != nil {
		f.InStack = *parent.InStack.Clone()
	}

	return f
}

func (f *Frame) Clone() *Frame {
	return &Frame{
		Parent:   f.Parent,
		InStack:  *f.InStack.Clone(),
		OutStack: *f.OutStack.Clone(),
	}
}
