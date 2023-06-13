package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/greenboxal/agibootstrap/pkg/codex"
)

func main() {
	var rootCmd = &cobra.Command{Use: "app"}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		Long:  "This command initializes a new project.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Initializing a new project...")
		},
	}

	var reindexCmd = &cobra.Command{ //Added reindex command
		Use:   "reindex",
		Short: "Reindex the project",
		Long:  "This command reindex the project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := os.Getwd()

			if err != nil {
				return err
			}

			cmd.SilenceUsage = true

			p, err := codex.NewProject(wd)

			if err != nil {
				return err
			}

			return p.Reindex()
		},
	}

	var generateCmd = &cobra.Command{
		Use:   "generate [repo path]",
		Short: "Generate a new file",
		Long:  "This command generates a new file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := os.Getwd()

			if err != nil {
				return err
			}

			if len(args) > 0 {
				wd = args[0]
			}

			cmd.SilenceUsage = true

			p, err := codex.NewProject(wd)

			if err != nil {
				return err
			}

			_, err = p.Generate()

			if err != nil {
				fmt.Printf("error: %s\n", err)
			}

			return err
		},
	}

	var commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "Commit current staged changes with automatic commit message.",
		Long:  "This command commits current staged changes with automatic commit message.",
		RunE: func(cmd *cobra.Command, args []string) error {
			wd, err := os.Getwd()

			if err != nil {
				return err
			}

			cmd.SilenceUsage = true

			p, err := codex.NewProject(wd)

			if err != nil {
				return err
			}

			return p.Commit()
		},
	}

	rootCmd.AddCommand(initCmd, reindexCmd, generateCmd, commitCmd) // Added reindexCmd in the command execution

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
