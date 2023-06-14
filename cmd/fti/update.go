package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:   "update [path]",
	Short: "Update the FTI repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Updating FTI at path:", args[0])
		// TODO: Add logic for update
	},
}
