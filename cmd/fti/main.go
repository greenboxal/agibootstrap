package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "fti",
		Short: "FTI is a tool for managing File Tree Index",
	}

	rootCmd.AddCommand(InitCmd, UpdateCmd, QueryCmd)

	err := rootCmd.Execute()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
