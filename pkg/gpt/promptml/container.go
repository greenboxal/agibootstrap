package promptml

import (
	"context"

	"github.com/greenboxal/aip/aip-langchain/pkg/tokenizers"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/iterators"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
)

type Parent interface {
	Node

	PmlParent() Parent
	PmlContainer() *Container

	GetTokenBuffer() *rendering.TokenBuffer

	IsNeedLayout() bool
	RequestLayout()

	Layout(ctx context.Context) error
	LayoutChildren(ctx context.Context) error
}

type Container struct {
	NodeBase

	tb *rendering.TokenBuffer
}

func NewContainer(tokenizer tokenizers.BasicTokenizer) *Container {
	c := &Container{
		tb: rendering.NewTokenBuffer(tokenizer, 0),
	}

	c.Init(c, "")

	return c
}

func (n *Container) PmlContainer() *Container               { return n }
func (n *Container) PmlParent() Parent                      { return n.PsiNode().(Parent) }
func (n *Container) GetTokenBuffer() *rendering.TokenBuffer { return n.tb }
func (n *Container) NeedsLayout() bool                      { return n.needsLayout }

func (n *Container) RequestLayout() {
	n.needsLayout = true

	n.Invalidate()
}

func (n *Container) Layout(ctx context.Context) error {
	n.RequestLayout()

	for !n.IsValid() || n.PmlContainer().NeedsLayout() {
		if err := n.Update(ctx); err != nil {
			return nil
		}
	}

	return nil
}

func (n *Container) LayoutChildren(ctx context.Context) error {
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

	if remaining > 0 {
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

	return nil
}

func (n *Container) Update(ctx context.Context) error {
	if n.NodeBase.Update(ctx) != nil {
		return nil
	}

	if n.needsLayout {
		if err := n.PmlParent().LayoutChildren(ctx); err != nil {
			return err
		}

		n.needsLayout = false
	}

	n.tb.Reset()

	for it := n.ChildrenIterator(); it.Next(); {
		leaf, isLeaf := it.Value().(Leaf)

		if !isLeaf {
			continue
		}

		if err := leaf.Render(ctx, n.tb); err != nil {
			return err
		}
	}

	n.currentLength = n.tb.TokenCount()

	return nil
}
