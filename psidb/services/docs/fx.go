package docs

import (
	"go.uber.org/fx"

	"github.com/greenboxal/agibootstrap/pkg/platform/inject"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering"
	"github.com/greenboxal/agibootstrap/psidb/psi/rendering/themes"
)

var Module = fx.Module(
	"services/docs",

	fx.Provide(NewIndexManager),

	inject.WithRegisteredService[*IndexManager](inject.ServiceRegistrationScopeSingleton),
)

func init() {
	rendering.WithSkinFunc(
		DocumentType,
		"text/markdown",
		"",
		func(ctx rendering.SkinRendererContext, node *Document) error {
			if _, err := ctx.Buffer.WriteFormat("# %s\n\n%s", node.Title, node.Content); err != nil {
				return err
			}

			return nil
		},
	)(themes.GlobalTheme.(*rendering.ThemeBase))
}
