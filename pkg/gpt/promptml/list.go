package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
)

type BuildListStreamFunc func(ctx context.Context) iterators.Iterator[Node]

type DynamicList struct {
	ContainerBase

	buildStream   BuildListStreamFunc
	currentStream iterators.Iterator[Node]
}

func NewDynamicList(streamBuilder BuildListStreamFunc) *DynamicList {
	l := &DynamicList{
		buildStream: streamBuilder,
	}

	l.Init(l)

	return l
}

func (l *DynamicList) LayoutChildren(ctx context.Context) error {
	if l.GetStage() != nil {
		if l.currentStream == nil {
			l.currentStream = l.buildStream(ctx)
		}

		currentLength := l.GetTokenLength()

		for currentLength < l.GetEffectiveMaxLength() && l.currentStream.Next() {
			child := l.currentStream.Value()

			l.AddChildNode(child)

			if err := child.Update(ctx); err != nil {
				return err
			}

			currentLength += child.GetTokenLength()
		}
	}

	return l.ContainerBase.LayoutChildren(ctx)
}
