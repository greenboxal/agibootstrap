package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/greenboxal/agibootstrap/pkg/fti"
)

var InitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new FTI repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()

		if err != nil {
			panic(err)
		}

		r, err := fti.NewRepository(cwd)

		if err != nil {
			panic(err)
		}

		return r.Init()
	},
}
