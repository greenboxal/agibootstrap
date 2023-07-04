package main

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/greenboxal/agibootstrap/pkg/build"
	"github.com/greenboxal/agibootstrap/pkg/build/codegen"
	"github.com/greenboxal/agibootstrap/pkg/build/fiximports"
	"github.com/greenboxal/agibootstrap/pkg/codex"
	"github.com/greenboxal/agibootstrap/pkg/visor"
	"github.com/greenboxal/agibootstrap/pkg/visor/chatui"

	// Register languages
	_ "github.com/greenboxal/agibootstrap/pkg/psi/langs/clang"
	_ "github.com/greenboxal/agibootstrap/pkg/psi/langs/golang"
	_ "github.com/greenboxal/agibootstrap/pkg/psi/langs/mdlang"
	_ "github.com/greenboxal/agibootstrap/pkg/psi/langs/pylang"
)

func main() {
	var rootCmd = &cobra.Command{Use: "app"}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		Long:  "This command initializes a new project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			wd, err := os.Getwd()

			if err != nil {
				return err
			}

			existingRoot, err := findProjectRoot()

			if err == nil && existingRoot != wd {
				return fmt.Errorf("a project already exists in this directory tree")
			}

			p, err := codex.NewBareProject(cmd.Context(), wd)

			if err != nil {
				return err
			}

			isValid, err := p.IsProjectValid(cmd.Context())

			if err != nil {
				return err
			}

			if isValid {
				return fmt.Errorf("a project already exists in this directory")
			}

			fmt.Println("Initializing a new project...")

			return p.Create(cmd.Context())
		},
	}

	var reindexCmd = &cobra.Command{
		Use:   "reindex",
		Short: "Reindex the project",
		Long:  "This command reindex the project.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			wd, err := findProjectRoot()

			if err != nil {
				return err
			}

			p, err := codex.LoadProject(cmd.Context(), wd)

			if err != nil {
				return err
			}

			defer p.Close()

			return p.Reindex()
		},
	}

	var generateCmd = &cobra.Command{
		Use:   "generate [repo path]",
		Short: "Generate a new file",
		Long:  "This command generates a new file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			wd, err := findProjectRoot()

			if err != nil {
				return err
			}

			if len(args) > 0 {
				wd = args[0]
			}

			p, err := codex.LoadProject(cmd.Context(), wd)

			if err != nil {
				return err
			}

			defer p.Close()

			builder := build.NewBuilder(p, build.Configuration{
				OutputDirectory: p.RootPath(),
				BuildDirectory:  path.Join(p.RootPath(), ".build"),

				BuildSteps: []build.Step{
					&codegen.BuildStep{},
					&fiximports.BuildStep{},
				},
			})

			_, err = builder.Build(cmd.Context())

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
			cmd.SilenceUsage = true

			wd, err := findProjectRoot()

			if err != nil {
				return err
			}

			p, err := codex.LoadProject(cmd.Context(), wd)

			if err != nil {
				return err
			}

			defer p.Close()

			return p.Commit()
		},
	}

	var debugCmd = &cobra.Command{
		Use:   "debug",
		Short: "Runs the debugger",
		Long:  "This command runs the debugger UI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			wd, err := findProjectRoot()

			if err != nil {
				return err
			}

			p, err := codex.LoadProject(cmd.Context(), wd)

			if err != nil {
				return err
			}

			defer p.Close()

			vis := visor.NewVisor(p)

			vis.Run()

			return nil
		},
	}

	var chatCmd = &cobra.Command{
		Use:   "chat",
		Short: "Runs the debugger chat",
		Long:  "This command runs the debugger chat UI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			wd, err := findProjectRoot()

			if err != nil {
				return err
			}

			p, err := codex.LoadProject(cmd.Context(), wd)

			if err != nil {
				return err
			}

			defer p.Close()

			vis := chatui.NewChatUI(p)

			vis.Run()

			return nil
		},
	}

	rootCmd.AddCommand(initCmd, reindexCmd, generateCmd, commitCmd, debugCmd, chatCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)

		os.Exit(1)
	}

	os.Exit(0)
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()

	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(path.Join(wd, ".codex")); err == nil {
			break
		}

		if _, err := os.Stat(path.Join(wd, "Codex.project.toml")); err == nil {
			break
		}

		if wd == "/" {
			return "", errors.New("could not find project root")
		}

		wd = path.Dir(wd)
	}

	return wd, nil
}
