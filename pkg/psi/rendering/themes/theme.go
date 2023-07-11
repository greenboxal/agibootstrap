package themes

import (
	"fmt"
	"html/template"
	"io"
	"net/url"

	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
)

var GenericTemplateHelpers = template.FuncMap{
	"psiNodePath": func(v any) psi.Path {
		return v.(psi.Node).CanonicalPath()
	},

	"psiPathEscape": func(v any) string {
		switch v := v.(type) {
		case psi.EdgeReference:
			return url.PathEscape(v.GetKey().String())

		case psi.PathElement:
			return url.PathEscape(v.String())

		case *psi.Path:
			return v.Format(true)

		case psi.Path:
			return v.Format(true)

		case string:
			return url.PathEscape(v)
		}

		panic(fmt.Errorf("invalid type %T", v))
	},
}

var GlobalTheme = rendering.BuildTheme(
	rendering.InheritTheme(rendering.DefaultTheme()),

	rendering.WithSkinFunc(
		vfs.FileType,
		"text/markdown",
		"",
		func(ctx rendering.SkinRendererContext, node *vfs.File) error {
			fh, err := node.Open()

			if err != nil {
				return err
			}

			defer fh.Close()

			reader, err := fh.Get()

			if err != nil {
				return err
			}

			defer reader.Close()

			if _, err := ctx.Buffer.WriteFormat("**%s:**```\n", node.Path()); err != nil {
				return err
			}

			if _, err := io.Copy(ctx.Buffer, reader); err != nil {
				return err
			}

			if _, err := ctx.Buffer.WriteString("\n```\n"); err != nil {
				return err
			}

			return nil
		},
	),

	rendering.WithSkinFunc(
		vfs.DirectoryType,
		"text/markdown",
		"",
		func(ctx rendering.SkinRendererContext, node *vfs.Directory) error {
			if _, err := ctx.Buffer.WriteFormat("**%s:**\n", node.Path()); err != nil {
				return err
			}

			for _, child := range node.Children() {
				fsNode, ok := child.(vfs.Node)

				if !ok {
					continue
				}

				if _, err := ctx.Buffer.WriteFormat("* %s\n", fsNode.Name()); err != nil {
					return err
				}
			}

			return nil
		},
	),
)
