package project

import (
	"context"

	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

var SourceFileEdge = psi.DefineEdgeType[SourceFile]("SourceFile", psi.WithEdgeTypeNamed())

func GetOrCreateSourceForFile(ctx context.Context, file *vfs.File, lang Language) (SourceFile, error) {
	key := SourceFileEdge.Named(lang.Name().String())
	existing, ok := psi.GetEdge[SourceFile](file, key)

	if ok && existing != nil {
		return existing, nil
	}

	fh, err := file.Open()

	if err != nil {
		return nil, err
	}

	src := lang.CreateSourceFile(ctx, file.Path, fh)

	src.SetParent(file)
	psi.UpdateEdge(file, key, src)

	return src, nil
}
