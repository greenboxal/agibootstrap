package codeanalysis

import (
	"context"
	"reflect"

	"github.com/greenboxal/agibootstrap/psidb/psi"
	"github.com/greenboxal/agibootstrap/psidb/typesystem"
)

type ISourceFile interface {
	Parse(ctx context.Context) error
}

type SourceFile struct {
	psi.NodeBase

	Origin   string `json:"origin"`
	Content  string `json:"content"`
	Language string `json:"language"`
}

var _ ISourceFile = (*SourceFile)(nil)
var SourceFileInterface = psi.DefineNodeInterface[ISourceFile]()
var SourceFileType = psi.DefineNodeType[*SourceFile](psi.WithInterfaceFromNode(SourceFileInterface))

func NewSourceFile(origin string, content string) *SourceFile {
	sf := &SourceFile{
		Origin:  origin,
		Content: content,
	}

	sf.Init(sf)

	return sf
}

func (sf *SourceFile) PsiNodeName() string { return sf.Origin }

func (sf *SourceFile) GetLanguage() Language {
	typ := typesystem.Universe().LookupByName(sf.Language)
	instance := reflect.New(typ.RuntimeType())

	return instance.Interface().(Language)
}

func (sf *SourceFile) Parse(ctx context.Context) error {
	lang := sf.GetLanguage()
	cu, err := lang.Parse(ctx, sf)

	if err != nil {
		return err
	}

	cu.GetRoot().SetParent(sf)

	return sf.Update(ctx)
}
