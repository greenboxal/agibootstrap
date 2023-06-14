package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/greenboxal/agibootstrap/pkg/fti"
)

var QueryCmd = &cobra.Command{
	Use:   "query [query terms]",
	Short: "Query the FTI repository",
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

		return r.Query(cmd.Context(), args[0])
	},
}

func init() {
	QueryCmd.Flags().String("query", "", "query terms for searching the FTI repository")
}
