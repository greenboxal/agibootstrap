package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
)

type Parent interface {
	Node

	PmlParent() Parent
	PmlContainer() *ContainerBase

	GetTokenBuffer() *rendering.TokenBuffer

	NeedsLayout() bool
	RequestLayout()

	Layout(ctx context.Context) error
	LayoutChildren(ctx context.Context) error
}

type ContainerBase struct {
	NodeBase

	isFirstLayout bool
}

func NewContainer() *ContainerBase {
	c := &ContainerBase{}

	c.Init(c, "")

	return c
}

func (n *ContainerBase) PmlContainer() *ContainerBase           { return n }
func (n *ContainerBase) PmlParent() Parent                      { return n.PsiNode().(Parent) }
func (n *ContainerBase) GetTokenBuffer() *rendering.TokenBuffer { return n.tb }
func (n *ContainerBase) NeedsLayout() bool                      { return n.needsLayout }

func (n *ContainerBase) RequestLayout() {
	n.needsLayout = true

	n.Invalidate()
}

func (n *ContainerBase) Layout(ctx context.Context) error {
	n.RequestLayout()

	for !n.IsValid() || n.PmlContainer().NeedsLayout() {
		if err := n.Update(ctx); err != nil {
			return nil
		}
	}

	return nil
}

func (n *ContainerBase) LayoutChildren(ctx context.Context) error {
	children := iterators.ToSlice(iterators.FilterIsInstance[psi.Node, Node](n.ChildrenIterator()))

	if len(children) == 0 {
		return nil
	}

	staticLength := 0
	dynamicLength := 0
	minLength := 0
	treeWeight := 0.0

	for _, child := range children {
		if child.IsResizable() {
			dynamicLength += child.GetTokenLength()
		} else {
			staticLength += child.GetTokenLength()
		}

		minLength += child.PmlNodeBase().GetEffectiveMinLength()
		treeWeight += child.GetBias()
	}

	totalLength := staticLength + dynamicLength
	remaining := n.GetEffectiveMaxLength() - totalLength

	if remaining > 0 && !n.isFirstLayout {
		return nil
	}

	remainingDynamicTokens := n.GetEffectiveMaxLength() - staticLength

	resizableOrdered := iterators.FilterIsInstance[psi.Node, Node](n.ChildrenIterator())

	resizableOrdered = iterators.Filter(resizableOrdered, func(n Node) bool {
		return n.IsResizable()
	})

	resizableOrdered = iterators.SortWith(resizableOrdered, func(a, b Node) int {
		aNorm := a.GetBias() / treeWeight
		bNorm := b.GetBias() / treeWeight

		return int(aNorm - bNorm)
	})

	for it := resizableOrdered; it.Next(); {
		child := it.Value()

		biasNorm := child.GetBias() / treeWeight

		child.SetMaxLength(NewTokenLength(float64(remainingDynamicTokens), TokenUnitToken))
		child.SetPrefLength(NewTokenLength(float64(remainingDynamicTokens)*biasNorm, TokenUnitToken))

		if err := child.Update(ctx); err != nil {
			return err
		}

		remainingDynamicTokens -= child.GetTokenLength()
	}

	n.isFirstLayout = false

	return nil
}

func (n *ContainerBase) Update(ctx context.Context) error {
	if n.needsLayout {
		if err := n.PmlParent().LayoutChildren(ctx); err != nil {
			return err
		}

		n.needsLayout = false
	}

	if n.NodeBase.Update(ctx) != nil {
		return nil
	}

	n.tb.Reset()

	for it := n.ChildrenIterator(); it.Next(); {
		leaf, isNode := it.Value().(Node)

		if !isNode {
			continue
		}

		tb := leaf.PmlNodeBase().GetTokenBuffer()

		if tb == nil {
			continue
		}

		_, err := n.tb.WriteBuffer(tb)

		if err != nil {
			return err
		}
	}

	return nil
}
