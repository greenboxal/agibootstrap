package rendering

import (
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type SkinRenderer interface {
	GetNodeType() psi.NodeType
	GetSkinName() string
	GetContentType() string

	RenderNode(ctx SkinRendererContext, node psi.Node) error
}

type Skin[T psi.Node] interface {
	SkinRenderer

	Render(ctx SkinRendererContext, node T) error
}

type SkinRendererContext struct {
	Renderer *PruningRenderer
	Theme    Theme
	Buffer   *TokenBuffer
}

type SkinFunc[T psi.Node] func(ctx SkinRendererContext, node T) error

type SkinBase[T psi.Node] struct {
	Name        string
	ContentType string
	NodeType    psi.NodeType

	RenderFn SkinFunc[T]
}

func (f SkinBase[T]) GetNodeType() psi.NodeType {
	return f.NodeType
}

func (f SkinBase[T]) GetSkinName() string    { return f.Name }
func (f SkinBase[T]) GetContentType() string { return f.ContentType }

func (f SkinBase[T]) Render(ctx SkinRendererContext, node T) error { return f.RenderFn(ctx, node) }
func (f SkinBase[T]) RenderNode(ctx SkinRendererContext, node psi.Node) error {
	return f.Render(ctx, node.(T))
}

type Theme interface {
	SkinForNode(contentType string, skinName string, node psi.Node) SkinRenderer
}

type SkinKey struct {
	NodeType    psi.NodeType
	ContentType string
	SkinName    string
}

type ThemeBase struct {
	Parent Theme
	Skins  map[SkinKey]SkinRenderer
}

func (t *ThemeBase) lookup(contentType, skinName string, node psi.Node) SkinRenderer {
	key := SkinKey{
		NodeType:    node.PsiNodeType(),
		ContentType: contentType,
		SkinName:    skinName,
	}

	return t.Skins[key]
}

func (t *ThemeBase) SkinForNode(contentType, skinName string, node psi.Node) SkinRenderer {
	skin := t.lookup(contentType, skinName, node)

	if skin != nil {
		return skin
	}

	if t.Parent != nil {
		if skin := t.Parent.SkinForNode(contentType, skinName, node); skin != nil {
			return skin
		}
	}

	return nil
}

type ThemeOption func(theme *ThemeBase)

func WithSkin(skin SkinRenderer) ThemeOption {
	return func(theme *ThemeBase) {
		key := SkinKey{
			NodeType:    skin.GetNodeType(),
			ContentType: skin.GetContentType(),
			SkinName:    skin.GetSkinName(),
		}

		theme.Skins[key] = skin
	}
}

func WithSkinFunc[T psi.Node](nodeType psi.TypedNodeType[T], contentType, skinName string, skin SkinFunc[T]) ThemeOption {
	key := SkinKey{
		NodeType:    nodeType,
		ContentType: contentType,
		SkinName:    skinName,
	}

	return func(theme *ThemeBase) {
		theme.Skins[key] = &SkinBase[T]{
			Name:        skinName,
			ContentType: contentType,
			RenderFn:    skin,
			NodeType:    nodeType,
		}
	}
}

func InheritTheme(parent Theme) ThemeOption {
	return func(theme *ThemeBase) {
		theme.Parent = parent
	}
}

func BuildTheme(options ...ThemeOption) Theme {
	theme := &ThemeBase{
		Skins: map[SkinKey]SkinRenderer{},
	}

	for _, option := range options {
		option(theme)
	}

	return theme
}

var defaultTheme = BuildTheme()

func DefaultTheme() Theme {
	return defaultTheme
}
