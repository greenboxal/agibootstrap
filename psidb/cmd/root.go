package cmd

import (
	"os"
	"path"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/greenboxal/agibootstrap/pkg/platform/api/apimachinery"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/workspace"
	"github.com/greenboxal/agibootstrap/psidb"
	coreapi "github.com/greenboxal/agibootstrap/psidb/core/api"
)

func BuildRoot(options ...fx.Option) *cobra.Command {
	extraOptions := fx.Options(options...)

	baseOptions := fx.Options(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logging.GetRootLogger()}
		}),

		psidb.BaseModules,
		workspace.Module,

		fx.Provide(func() (*coreapi.Config, error) {
			wd, err := os.Getwd()

			if err != nil {
				return nil, err
			}

			cfg := &coreapi.Config{}
			cfg.RootUUID = "QmYXZ"
			cfg.ProjectDir = wd
			cfg.DataDir = path.Join(cfg.ProjectDir, ".fti/psi")
			cfg.UseTLS = true

			cfg.Workers.MaxWorkers = 32
			cfg.Workers.MaxCapacity = 1024

			return cfg, nil
		}),

		extraOptions,
	)

	cmd := &cobra.Command{
		Use: "psidb",
	}

	cmd.AddCommand(buildServerCmd(baseOptions))

	return cmd
}

func buildServerCmd(options fx.Option) *cobra.Command {
	return &cobra.Command{
		Use: "server",

		RunE: func(cmd *cobra.Command, args []string) error {
			app := fx.New(
				options,

				fx.Invoke(func(server *apimachinery.Server) {}),
				fx.Invoke(func(wrk *workspace.Workspace) {}),
			)

			app.Run()

			return nil
		},
	}
}
