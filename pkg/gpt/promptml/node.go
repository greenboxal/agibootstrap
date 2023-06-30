package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type Node interface {
	psi.Node

	PmlNodeBase() *NodeBase
	PmlNode() Node

	GetLayoutBounds() Bounds
	GetBoundsInLocal() Bounds
	GetBoundsInParent() Bounds

	IsResizable() bool
	SetResizable(resizable bool)

	IsVisible() bool
	SetVisible(visible bool)

	IsMovable() bool
	SetMovable(movable bool)

	GetMinLength() TokenLength
	SetMinLength(length TokenLength)

	GetMaxLength() TokenLength
	SetMaxLength(length TokenLength)

	GetPrefLength() TokenLength
	SetPrefLength(length TokenLength)

	GetBias() float64
	SetBias(bias float64)

	GetTokenLength() int
}

type NodeBase struct {
	psi.NodeBase

	Visible bool `json:"visible,omitempty"`

	LayoutBounds *Bounds `json:"layout_bounds,omitempty"`

	Resizable bool `json:"resizable"`
	Movable   bool `json:"movable"`

	MinLength  TokenLength `json:"min_length"`
	MaxLength  TokenLength `json:"max_length"`
	PrefLength TokenLength `json:"pref_length"`

	Bias float64 `json:"bias"`

	needsLayout    bool
	boundsInParent *Bounds

	effectiveMinLength  int
	effectiveMaxLength  int
	effectivePrefLength int
	currentLength       int
}

func (n *NodeBase) PmlNodeBase() *NodeBase { return n }
func (n *NodeBase) PmlNode() Node          { return n.PsiNode().(Node) }
func (n *NodeBase) IsContainer() bool      { return true }
func (n *NodeBase) IsLeaf() bool           { return false }

func (n *NodeBase) GetLayoutBounds() Bounds {
	if n.LayoutBounds == nil {
		return n.GetBoundsInLocal()
	}

	return *n.LayoutBounds
}

func (n *NodeBase) GetBoundsInLocal() Bounds {
	return Bounds{
		Position: 0,
		Length:   NewTokenLength(float64(n.GetTokenLength()), TokenUnitToken),
	}
}

func (n *NodeBase) GetBoundsInParent() Bounds {
	if n.boundsInParent == nil {
		return n.GetLayoutBounds()
	}

	return *n.boundsInParent
}

func (n *NodeBase) GetRelevance() float64 { return 1 + n.Bias }
func (n *NodeBase) GetBias() float64      { return n.Bias }
func (n *NodeBase) SetBias(bias float64) {
	n.Bias = bias

	n.Invalidate()
	n.RequestParentLayout()
}

func (n *NodeBase) IsResizable() bool { return n.Resizable }
func (n *NodeBase) SetResizable(resizable bool) {
	n.Resizable = resizable
	n.Invalidate()
}

func (n *NodeBase) IsVisible() bool { return n.Visible }
func (n *NodeBase) SetVisible(visible bool) {
	n.Visible = visible
	n.Invalidate()
	n.RequestParentLayout()
}

func (n *NodeBase) IsMovable() bool { return n.Movable }
func (n *NodeBase) SetMovable(movable bool) {
	n.Movable = movable
	n.Invalidate()
	n.RequestParentLayout()
}

func (n *NodeBase) GetMinLength() TokenLength { return n.MinLength }
func (n *NodeBase) SetMinLength(length TokenLength) {
	n.MinLength = length
	n.Invalidate()
	n.RequestParentLayout()
}

func (n *NodeBase) GetMaxLength() TokenLength { return n.MaxLength }
func (n *NodeBase) SetMaxLength(length TokenLength) {
	n.MaxLength = length
	n.Invalidate()
	n.RequestParentLayout()
	n.PmlNodeBase().RequestParentLayout()
}

func (n *NodeBase) GetPrefLength() TokenLength { return n.PrefLength }
func (n *NodeBase) SetPrefLength(length TokenLength) {
	n.PrefLength = length
	n.Invalidate()
	n.RequestParentLayout()
}

func (n *NodeBase) GetEffectiveMaxLength() int  { return n.effectiveMaxLength }
func (n *NodeBase) GetEffectiveMinLength() int  { return n.effectiveMinLength }
func (n *NodeBase) GetEffectivePrefLength() int { return n.effectivePrefLength }

func (n *NodeBase) GetTokenLength() int { return n.currentLength }

func (n *NodeBase) RequestParentLayout() {
	if n.Parent() == nil {
		return
	}

	p, ok := n.Parent().(Parent)

	if !ok {
		return
	}

	p.RequestLayout()
}

func (n *NodeBase) Update(ctx context.Context) error {
	if err := n.NodeBase.Update(ctx); err != nil {
		return err
	}

	n.effectiveMinLength = n.MinLength.GetEffectiveLength(func(f float64) int {
		p := n.PmlParent()

		if p == nil {
			return -1
		}

		return p.PmlNodeBase().GetEffectiveMinLength()
	})

	n.effectiveMaxLength = n.MaxLength.GetEffectiveLength(func(f float64) int {
		p := n.PmlParent()

		if p == nil {
			return -1
		}

		return p.PmlNodeBase().GetEffectiveMaxLength()
	})

	n.effectivePrefLength = n.PrefLength.GetEffectiveLength(func(f float64) int {
		p := n.PmlParent()

		if p == nil {
			return -1
		}

		return p.PmlNodeBase().GetEffectivePrefLength()
	})

	if n.effectiveMaxLength < n.effectiveMinLength {
		n.effectiveMaxLength = n.effectiveMinLength
	}

	if n.effectivePrefLength < n.effectiveMinLength {
		n.effectivePrefLength = n.effectiveMinLength
	}

	if n.effectivePrefLength > n.effectiveMaxLength {
		n.effectivePrefLength = n.effectiveMaxLength
	}

	return nil
}

func (n *NodeBase) PmlParent() Parent {
	if n.Parent() == nil {
		return nil
	}

	p, ok := n.Parent().(Parent)

	if !ok {
		return nil
	}

	return p
}
