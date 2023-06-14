package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	main2 "github.com/greenboxal/agibootstrap/cmd/fti"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "fti",
		Short: "FTI is a tool for managing File Tree Index",
	}

	rootCmd.AddCommand(main2.InitCmd, main2.UpdateCmd, main2.QueryCmd)

	err := rootCmd.Execute()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
