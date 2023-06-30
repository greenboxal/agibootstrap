package promptml

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx"
	"github.com/greenboxal/agibootstrap/pkg/platform/stdlib/obsfx/collectionsfx"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
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

	GetRelevance() float64
	GetBias() float64
	SetBias(bias float64)

	GetTokenLength() int
	GetTokenLengthProperty() obsfx.ObservableValue[int]
}

type NodeBase struct {
	psi.NodeBase

	LayoutBounds obsfx.Property[*Bounds] `json:"layout_bounds,omitempty"`

	Visible   obsfx.BoolProperty `json:"visible,omitempty"`
	Resizable obsfx.BoolProperty `json:"resizable"`
	Movable   obsfx.BoolProperty `json:"movable"`

	MinLength obsfx.SimpleProperty[TokenLength] `json:"min_length"`
	MaxLength obsfx.SimpleProperty[TokenLength] `json:"max_length"`

	Bias obsfx.DoubleProperty `json:"bias"`

	effectiveMinLength obsfx.ObservableValue[int]
	effectiveMaxLength obsfx.ObservableValue[int]
	relevance          obsfx.DoubleProperty

	tokenLength obsfx.IntProperty

	stage obsfx.SimpleProperty[*Stage]

	tb *rendering.TokenBuffer

	boundsInParent *Bounds
}

func (n *NodeBase) PmlNodeBase() *NodeBase { return n }
func (n *NodeBase) PmlNode() Node          { return n.PsiNode().(Node) }

func (n *NodeBase) GetStage() *Stage {
	s := n.stage.Value()

	if s == nil {
		if p := n.PmlParent(); p != nil {
			return p.PmlNodeBase().GetStage()
		}
	}

	return n.stage.Value()
}
func (n *NodeBase) GetTokenBuffer() *rendering.TokenBuffer { return n.tb }

func (n *NodeBase) GetTokenLength() int                                { return n.tokenLength.Value() }
func (n *NodeBase) GetTokenLengthProperty() obsfx.ObservableValue[int] { return &n.tokenLength }

func (n *NodeBase) IsContainer() bool { return true }
func (n *NodeBase) IsLeaf() bool      { return false }

func (n *NodeBase) GetEffectiveMinLength() int {
	v := n.MinLength.Value()

	if !v.IsReal() {
		if p := n.PmlParent(); p != nil {
			v = v.MulInt(p.PmlNodeBase().GetEffectiveMinLength())
		} else {
			return 0
		}
	}

	return v.TokenCount()
}

func (n *NodeBase) GetEffectiveMaxLength() int {
	v := n.MaxLength.Value()

	if !v.IsReal() {
		if p := n.PmlParent(); p != nil {
			pml := p.PmlNodeBase().GetEffectiveMaxLength()

			v = v.MulInt(pml)
		} else if n.GetStage() != nil {
			v = v.MulInt(n.GetStage().MaxTokens)
		}
	}

	return v.TokenCount()
}

func (n *NodeBase) Init(self psi.Node, uuid string) {
	n.MinLength.SetValue(NewTokenLength(0, TokenUnitPercent))
	n.MaxLength.SetValue(NewTokenLength(1, TokenUnitPercent))
	n.Bias.SetValue(0.0)
	n.Visible.SetValue(true)
	n.Resizable.SetValue(true)
	n.Movable.SetValue(true)

	n.relevance.Bind(obsfx.BindExpression(func() float64 {
		return 1.0 + n.Bias.Value()
	}, &n.Bias))

	n.NodeBase.Init(self, uuid)

	obsfx.ObserveInvalidation(&n.stage, func() {
		for it := n.ChildrenIterator(); it.Next(); {
			cn, ok := it.Value().(Node)

			if !ok {
				continue
			}

			cn.PmlNodeBase().setStage(n.stage.Value())
		}

		n.Invalidate()
	})

	n.effectiveMinLength = obsfx.BindExpression(func() int {
		return n.GetEffectiveMinLength()
	}, &n.MaxLength, &n.MinLength, n.ParentProperty(), &n.stage)

	n.effectiveMaxLength = obsfx.BindExpression(func() int {
		return n.GetEffectiveMaxLength()
	}, &n.MaxLength, &n.MinLength, n.ParentProperty(), &n.stage)

	obsfx.ObserveInvalidation(n.effectiveMinLength, n.InvalidateLayout)
	obsfx.ObserveInvalidation(n.effectiveMaxLength, n.InvalidateLayout)

	obsfx.ObserveInvalidation(n.PsiNodeBase().ParentProperty(), n.InvalidateLayout)
	obsfx.ObserveInvalidation(&n.MinLength, n.InvalidateLayout)
	obsfx.ObserveInvalidation(&n.MaxLength, n.InvalidateLayout)
	obsfx.ObserveInvalidation(&n.Bias, n.InvalidateLayout)
	obsfx.ObserveInvalidation(&n.Visible, n.InvalidateLayout)
	obsfx.ObserveInvalidation(&n.Resizable, n.InvalidateLayout)
	obsfx.ObserveInvalidation(&n.Movable, n.InvalidateLayout)

	collectionsfx.ObserveList(n.ChildrenList(), func(ev collectionsfx.ListChangeEvent[psi.Node]) {
		if ev.WasAdded() {
			for _, child := range ev.AddedSlice() {
				cn, ok := child.(Node)

				if !ok {
					continue
				}

				if n.GetStage() != nil {
					cn.PmlNodeBase().setStage(n.GetStage())
				}
			}
		}
	})
}

func (n *NodeBase) GetLayoutBounds() Bounds {
	if n.LayoutBounds == nil {
		return n.GetBoundsInLocal()
	}

	return *n.LayoutBounds.Value()
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

func (n *NodeBase) RelevanceProperty() obsfx.Property[float64] { return &n.relevance }
func (n *NodeBase) GetRelevance() float64                      { return n.relevance.Value() }
func (n *NodeBase) GetBias() float64                           { return n.Bias.Value() }
func (n *NodeBase) SetBias(bias float64)                       { n.Bias.SetValue(bias) }
func (n *NodeBase) IsResizable() bool                          { return n.Resizable.Value() }
func (n *NodeBase) SetResizable(resizable bool)                { n.Resizable.SetValue(resizable) }
func (n *NodeBase) IsVisible() bool                            { return n.Visible.Value() }
func (n *NodeBase) SetVisible(visible bool)                    { n.Visible.SetValue(visible) }
func (n *NodeBase) IsMovable() bool                            { return n.Movable.Value() }
func (n *NodeBase) SetMovable(movable bool)                    { n.Movable.SetValue(movable) }
func (n *NodeBase) GetMinLength() TokenLength                  { return n.MinLength.Value() }
func (n *NodeBase) SetMinLength(length TokenLength)            { n.MinLength.SetValue(length) }
func (n *NodeBase) GetMaxLength() TokenLength                  { return n.MaxLength.Value() }
func (n *NodeBase) SetMaxLength(length TokenLength)            { n.MaxLength.SetValue(length) }

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

func (n *NodeBase) updateDimensions() {
	if n.GetStage() != nil {
		if n.tb == nil {
			n.tb = rendering.NewTokenBuffer(n.GetStage().Tokenizer, n.GetEffectiveMaxLength())
		} else {
			n.tb.SetTokenLimit(n.GetEffectiveMaxLength())
		}
	}
}

func (n *NodeBase) Update(ctx context.Context) error {
	if p := n.PmlParent(); p != nil {
		ps := p.PmlNodeBase().GetStage()

		if ps != nil && n.GetStage() != ps {
			n.setStage(ps)
		}
	}

	n.updateDimensions()

	if err := n.NodeBase.Update(ctx); err != nil {
		return err
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

func (n *NodeBase) setStage(stage *Stage) {
	n.stage.SetValue(stage)
}

func (n *NodeBase) InvalidateLayout() {
	n.Invalidate()
	n.RequestParentLayout()
}
