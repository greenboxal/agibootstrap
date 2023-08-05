package fiximports

import (
	"context"
	"strings"

	"golang.org/x/tools/imports"

	build2 "github.com/greenboxal/agibootstrap/pkg/legacy/build"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type BuildStep struct{}

func (s *BuildStep) Process(ctx context.Context, bctx *build2.Context) (result build2.StepResult, err error) {
	err = psi.Walk(bctx.Project(), func(cursor psi.Cursor, entering bool) error {
		n := cursor.Value()

		cursor.SkipChildren()

		if entering {
			switch n := n.(type) {
			case project.Project:
				cursor.WalkChildren()

			case *vfs.Directory:
				cursor.WalkChildren()

			case *vfs.File:
				if !strings.HasSuffix(n.GetPath(), ".go") {
					break
				}

				opt := &imports.Options{
					FormatOnly: false,
					AllErrors:  true,
					Comments:   true,
					TabIndent:  true,
					TabWidth:   4,
					Fragment:   false,
				}

				sf, err := bctx.Project().GetSourceFile(ctx, n.GetPath())

				if err != nil {
					return err
				}

				code, err := sf.ToCode(sf.Root())

				if err != nil {
					return err
				}

				newCode, err := imports.Process(n.GetPath(), []byte(code.Code), opt)

				if err != nil {
					return err
				}

				if string(newCode) != code.Code {
					err = sf.Replace(ctx, string(newCode))

					if err != nil {
						return err
					}

					result.ChangeCount++
				}
			}
		}

		return nil
	})

	return result, nil
}
