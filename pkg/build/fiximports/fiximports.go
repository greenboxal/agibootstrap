package fiximports

import (
	"context"
	"strings"

	"golang.org/x/tools/imports"

	"github.com/greenboxal/agibootstrap/pkg/build"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

type BuildStep struct{}

func (s *BuildStep) Process(ctx context.Context, bctx *build.Context) (result build.StepResult, err error) {
	err = psi.Walk(bctx.Project(), func(cursor psi.Cursor, entering bool) error {
		n := cursor.Node()

		cursor.SkipChildren()

		if entering {
			switch n := n.(type) {
			case project.Project:
				cursor.WalkChildren()

			case *vfs.DirectoryNode:
				cursor.WalkChildren()

			case *vfs.FileNode:
				if !strings.HasSuffix(n.Path(), ".go") {
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

				sf, err := bctx.Project().GetSourceFile(n.Path())

				if err != nil {
					return err
				}

				code, err := sf.ToCode(sf.Root())

				if err != nil {
					return err
				}

				newCode, err := imports.Process(n.Path(), []byte(code.Code), opt)

				if err != nil {
					return err
				}

				if string(newCode) != code.Code {
					err = sf.Replace(string(newCode))

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
