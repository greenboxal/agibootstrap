package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 'query' command
var QueryCmd = &cobra.Command{
	Use:   "query [path] --query [query terms]",
	Short: "Query the FTI repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		queryTerms, _ := cmd.Flags().GetString("query")
		fmt.Println("Querying FTI at path:", args[0], "with query terms:", queryTerms)
		// TODO: Add logic for query
	},
}

func init() {
	QueryCmd.Flags().String("query", "", "query terms for searching the FTI repository")
}
