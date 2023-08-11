package rendering

import (
	"context"
	"net/url"
	"reflect"

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
	Context context.Context
	Query   url.Values

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
	NodeType    string
	ContentType string
	SkinName    string
}

type SkinKeyPattern struct {
	NodeType    *psi.NodeType
	ContentType *string
	SkinName    *string
	Renderer    SkinRenderer

	NodeFilterFn func(node psi.Node, contentType, skinName string) bool
}

func (skp SkinKeyPattern) MatchKey(key SkinKey) bool {
	if skp.NodeType != nil && (*skp.NodeType).Name() != key.NodeType {
		return false
	}

	if skp.ContentType != nil && *skp.ContentType != key.ContentType {
		return false
	}

	if skp.SkinName != nil && *skp.SkinName != key.SkinName {
		return false
	}

	return true
}

func (skp SkinKeyPattern) MatchNode(n psi.Node, contentType, skinName string) bool {
	if skp.NodeType != nil && *skp.NodeType != n.PsiNodeType() {
		return false
	}

	if skp.ContentType != nil && *skp.ContentType != contentType {
		return false
	}

	if skp.SkinName != nil && *skp.SkinName != skinName {
		return false
	}

	if skp.NodeFilterFn != nil && !skp.NodeFilterFn(n, contentType, skinName) {
		return false
	}

	return true
}

type ThemeBase struct {
	Parent Theme

	Skins    map[SkinKey]SkinRenderer
	Patterns []SkinKeyPattern
}

func (t *ThemeBase) lookup(contentType, skinName string, node psi.Node) SkinRenderer {
	key := SkinKey{
		NodeType:    node.PsiNodeType().Name(),
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

	for _, pattern := range t.Patterns {
		if pattern.MatchNode(node, contentType, skinName) {
			return pattern.Renderer
		}
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
			NodeType:    skin.GetNodeType().Name(),
			ContentType: skin.GetContentType(),
			SkinName:    skin.GetSkinName(),
		}

		theme.Skins[key] = skin
	}
}

func WithSkinFunc[T psi.Node](nodeType psi.TypedNodeType[T], contentType, skinName string, skin SkinFunc[T]) ThemeOption {
	key := SkinKey{
		NodeType:    nodeType.Name(),
		ContentType: contentType,
		SkinName:    skinName,
	}

	sb := &SkinBase[T]{
		Name:        skinName,
		ContentType: contentType,
		RenderFn:    skin,
		NodeType:    nodeType,
	}

	return func(theme *ThemeBase) {
		theme.Skins[key] = sb
	}
}

func WithSuperclassSkinFunc[Super psi.Node](contentType, skinName string, skin SkinFunc[Super]) ThemeOption {
	superTyp := reflect.TypeOf((*Super)(nil)).Elem()

	sb := &SkinBase[Super]{
		Name:        skinName,
		ContentType: contentType,
		RenderFn:    skin,
	}

	return WithSkinPattern(SkinKeyPattern{
		ContentType: &contentType,
		SkinName:    &skinName,
		Renderer:    sb,

		NodeFilterFn: func(node psi.Node, contentType, skinName string) bool {
			return reflect.ValueOf(node).Type().AssignableTo(superTyp)
		},
	})
}

func WithSkinPattern(pattern SkinKeyPattern) ThemeOption {
	return func(theme *ThemeBase) {
		theme.Patterns = append(theme.Patterns, pattern)
	}
}

func WithSkinPatternFunc[T psi.Node](nodeType *psi.TypedNodeType[T], contentType, skinName *string, skin SkinFunc[T]) ThemeOption {
	key := SkinKeyPattern{
		ContentType: contentType,
		SkinName:    skinName,
	}

	if nodeType != nil {
		nt := psi.NodeType(*nodeType)
		key.NodeType = &nt
	}

	sb := &SkinBase[T]{
		RenderFn: skin,
	}

	if contentType != nil {
		sb.ContentType = *contentType
	}

	if skinName != nil {
		sb.Name = *skinName
	}

	if key.NodeType != nil {
		sb.NodeType = *key.NodeType
	}

	key.Renderer = sb

	return WithSkinPattern(key)
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
