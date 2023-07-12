package main

import (
	"context"
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

var projectRoot string
var project *codex.Project

func initializeProject(ctx context.Context, load bool) error {
	if projectRoot == "" {
		root, err := findProjectRoot()

		if err == nil {
			projectRoot = root
		}
	}

	if projectRoot != "" {
		p, err := codex.NewBareProject(projectRoot)

		if err != nil {
			return err
		}

		project = p
	}

	if project != nil && load {
		if err := project.Load(ctx); err != nil {
			return err
		}
	}

	return nil
}

func teardownProject(ctx context.Context) error {
	if project != nil {
		if err := project.Shutdown(ctx); err != nil {
			return err
		}

		project = nil
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use: "app",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeProject(cmd.Context(), false)
		},

		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return teardownProject(cmd.Context())
		},
	}

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

			if projectRoot != "" && projectRoot != wd {
				return fmt.Errorf("a project already exists in this directory tree")
			}

			isValid, err := project.IsProjectValid(cmd.Context())

			if err != nil {
				return err
			}

			if isValid {
				return fmt.Errorf("a project already exists in this directory")
			}

			fmt.Println("Initializing a new project...")

			return project.Create(cmd.Context())
		},
	}

	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a new file",
		Long:  "This command generates a new file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := project.Load(cmd.Context()); err != nil {
				return err
			}

			builder := build.NewBuilder(project, build.Configuration{
				OutputDirectory: project.RootPath(),
				BuildDirectory:  path.Join(project.RootPath(), ".build"),

				BuildSteps: []build.Step{
					&codegen.BuildStep{},
					&fiximports.BuildStep{},
				},
			})

			_, err := builder.Build(cmd.Context())

			if err != nil {
				fmt.Printf("error: %s\n", err)
			}

			return err
		},
	}

	var analyzeCmd = &cobra.Command{
		Use:   "analyze [path]",
		Short: "Analyze a specific file or directory",
		Long:  "This command analyzes a specific file or directory.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := project.Load(cmd.Context()); err != nil {
				return err
			}

			/*targetPath := project.CanonicalPath()

			if len(args) > 0 {
				p, err := psi.ParsePath(args[0])

				if err != nil {
					return err
				}

				targetPath = p
			}*/

			return project.Reindex()
		},
	}

	var commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "Commit current staged changes with automatic commit message.",
		Long:  "This command commits current staged changes with automatic commit message.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := project.Load(cmd.Context()); err != nil {
				return err
			}

			return project.Commit()
		},
	}

	var debugCmd = &cobra.Command{
		Use:   "debug",
		Short: "Runs the debugger",
		Long:  "This command runs the debugger UI.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if err := project.Load(cmd.Context()); err != nil {
				return err
			}

			vis := visor.NewVisor(project)

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

			if err := project.Load(cmd.Context()); err != nil {
				return err
			}

			vis := chatui.NewChatUI(project)

			vis.Run()

			return nil
		},
	}

	psiDbCmd := buildPsiDbCmd()

	rootCmd.AddCommand(initCmd, analyzeCmd, generateCmd, commitCmd, debugCmd, chatCmd, psiDbCmd)

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
