package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new FTI repository",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		fmt.Println("Initializing FTI at path:", path)

		// Create .fti directory
		err := os.MkdirAll(path+"/.fti/objects", os.ModePerm)
		if err != nil {
			fmt.Println("Failed to create .fti/objects directory:", err)
			os.Exit(1)
		}

		// Create .fti/index directory
		err = os.MkdirAll(path+"/.fti/index", os.ModePerm)
		if err != nil {
			fmt.Println("Failed to create .fti/index directory:", err)
			os.Exit(1)
		}

		fmt.Println("FTI initialized successfully at path:", path)
	},
}
