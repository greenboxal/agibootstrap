package main

import (
	"fmt"
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

		hits, err := r.Query(cmd.Context(), args[0], 10)

		if err != nil {
			panic(err)
		}

		for i, hit := range hits {
			fmt.Printf("+ Hit %d (score = %f, ci = %d):\n%s\n", i, hit.Distance, hit.Entry.Chunk.Index, hit.Entry.Chunk.Content)
		}

		return nil
	},
}

func init() {
	QueryCmd.Flags().String("query", "", "query terms for searching the FTI repository")
}
