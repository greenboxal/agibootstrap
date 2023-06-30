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

	isLayoutValid bool
}

func NewContainer() *ContainerBase {
	c := &ContainerBase{}

	c.Init(c, "")

	return c
}

func (n *ContainerBase) PmlContainer() *ContainerBase           { return n }
func (n *ContainerBase) AsPmlParent() Parent                    { return n.PsiNode().(Parent) }
func (n *ContainerBase) GetTokenBuffer() *rendering.TokenBuffer { return n.tb }
func (n *ContainerBase) NeedsLayout() bool                      { return !n.isLayoutValid }

func (n *ContainerBase) RequestLayout() {
	n.isLayoutValid = false

	n.Invalidate()
}

func (n *ContainerBase) Layout(ctx context.Context) error {
	n.RequestLayout()

	return n.Update(ctx)
}

func (n *ContainerBase) LayoutChildren(ctx context.Context) error {
	children := iterators.ToSlice(iterators.FilterIsInstance[psi.Node, Node](n.ChildrenIterator()))

	if len(children) == 0 {
		return nil
	}

	staticLength := 0
	dynamicLength := 0
	treeWeight := 0.0

	for _, child := range children {
		if err := child.Update(ctx); err != nil {
			return err
		}

		if child.IsResizable() {
			dynamicLength += child.GetTokenLength()
			treeWeight += child.PmlNodeBase().GetRelevance()
		} else {
			staticLength += child.GetTokenLength()
		}
	}

	maxDynamicTokens := n.GetEffectiveMaxLength() - staticLength
	remainingDynamicTokens := maxDynamicTokens

	resizableOrdered := iterators.FilterIsInstance[psi.Node, Node](n.ChildrenIterator())

	resizableOrdered = iterators.Filter(resizableOrdered, func(n Node) bool {
		return n.IsResizable()
	})

	resizableOrdered = iterators.SortWith(resizableOrdered, func(a, b Node) int {
		aNorm := a.PmlNodeBase().GetRelevance() / treeWeight
		bNorm := b.PmlNodeBase().GetRelevance() / treeWeight

		return int(bNorm - aNorm)
	})

	tempBuffer := rendering.NewTokenBuffer(n.GetStage().Tokenizer, remainingDynamicTokens)

	for it := resizableOrdered; it.Next(); {
		child := it.Value()

		biasNorm := child.PmlNodeBase().GetRelevance() / treeWeight

		if child.GetMinLength().Unit != TokenUnitPercent {
			child.SetMinLength(NewTokenLength(float64(remainingDynamicTokens)*biasNorm, TokenUnitToken))
		}

		if child.GetMaxLength().Unit != TokenUnitPercent {
			child.SetMaxLength(NewTokenLength(float64(remainingDynamicTokens), TokenUnitToken))
		}

		if err := child.Update(ctx); err != nil {
			return err
		}

		childTb := child.PmlNodeBase().GetTokenBuffer()

		if childTb != nil {
			if _, err := childTb.WriteTo(tempBuffer); err != nil {
				return err
			}

			remainingDynamicTokens = maxDynamicTokens - tempBuffer.TokenCount()
		}
	}

	return nil
}

func (n *ContainerBase) Update(ctx context.Context) error {
	if n.GetStage() == nil {
		return nil
	}

	if err := n.NodeBase.Update(ctx); err != nil {
		return err
	}

	if !n.isLayoutValid {
		if err := n.AsPmlParent().LayoutChildren(ctx); err != nil {
			return err
		}

		for it := n.ChildrenIterator(); it.Next(); {
			if err := it.Node().Update(ctx); err != nil {
				return err
			}
		}

		n.isLayoutValid = true
	}

	if n.tb != nil {
		n.tb.Reset()

		if err := n.render(ctx, n.tb); err != nil {
			return err
		}

		n.tokenLength.SetValue(n.tb.TokenCount())
	}

	return nil
}

func (n *ContainerBase) render(ctx context.Context, tb *rendering.TokenBuffer) error {
	for it := n.ChildrenIterator(); it.Next(); {
		cn, isNode := it.Value().(Node)

		if !isNode {
			continue
		}

		if !cn.IsVisible() {
			continue
		}

		childTb := cn.PmlNodeBase().GetTokenBuffer()

		if childTb == nil {
			continue
		}

		_, err := childTb.WriteTo(tb)

		if err != nil {
			return err
		}

	}

	return nil
}
