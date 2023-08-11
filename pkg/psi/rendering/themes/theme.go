package themes

import (
	"fmt"
	"html/template"
	"io"
	"net/url"
	"path"

	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/rendering"
	"github.com/greenboxal/agibootstrap/psidb/services/kb"
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

			if _, err := ctx.Buffer.WriteFormat("```%s\n", path.Ext(node.GetPath())); err != nil {
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
			if _, err := ctx.Buffer.WriteFormat("**%s:**\n", node.GetPath()); err != nil {
				return err
			}

			for _, child := range node.Children() {
				fsNode, ok := child.(vfs.Node)

				if !ok {
					continue
				}

				if _, err := ctx.Buffer.WriteFormat("* %s\n", fsNode.GetName()); err != nil {
					return err
				}
			}

			return nil
		},
	),

	rendering.WithSuperclassSkinFunc[project.AstNode](
		"text/markdown",
		"",
		func(ctx rendering.SkinRendererContext, node project.AstNode) error {
			src := node.GetSourceFile()
			code, err := src.ToCode(node)

			if err != nil {
				return err
			}

			if node == src.Root() {
				_, err = fmt.Fprintf(ctx.Buffer, "**%s:**\n```%s\n%s\n```\n", code.Filename, code.Language, code.Code)
			} else {
				_, err = fmt.Fprintf(ctx.Buffer, "**%s (partial):**\n```%s\n%s\n```\n", code.Filename, code.Language, code.Code)
			}

			return err
		},
	),

	rendering.WithSkinFunc(
		kb.KnowledgeBaseType,
		"text/markdown",
		"",
		func(ctx rendering.SkinRendererContext, node *kb.KnowledgeBase) error {
			if _, err := ctx.Buffer.WriteFormat("# %s\n TODO", node.Name); err != nil {
				return err
			}

			return nil
		},
	),

	rendering.WithSkinFunc(
		kb.DocumentType,
		"text/markdown",
		"",
		func(ctx rendering.SkinRendererContext, node *kb.Document) error {
			if _, err := ctx.Buffer.WriteFormat("# %s\n%s\n", node.Title, node.Body); err != nil {
				return err
			}

			return nil
		},
	),

	rendering.WithSkinFunc(
		kb.DocumentType,
		"text/markdown",
		"summary",
		func(ctx rendering.SkinRendererContext, node *kb.Document) error {
			if _, err := ctx.Buffer.WriteFormat("# %s\n%s\n", node.Title, node.Summary); err != nil {
				return err
			}

			return nil
		},
	),
)
