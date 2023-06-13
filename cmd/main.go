package main

import (
	"bytes"
	"fmt"
	fs "io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/greenboxal/agibootstrap/pkg/codex"
	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/io"
	"github.com/greenboxal/agibootstrap/pkg/psi"
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

			_, err = proceseRepo(wd)

			if err != nil {
				return err
			}

			return nil
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

			fsRoot, err := NewFS(wd)

			if err != nil {
				return err
			}

			isDirty, err := fsRoot.IsDirty()

			if err != nil {
				return err
			}

			if !isDirty {
				return nil
			}

			diff, err := fsRoot.GetUncommittedChanges()

			if err != nil {
				return err
			}

			commitMessage, err := gpt.PrepareCommitMessage(diff)

			if err != nil {
				return err
			}

			commitId, err := fsRoot.Commit(commitMessage)

			if err != nil {
				return err
			}

			fmt.Printf("Changes committed with commit ID %s\n", commitId)

			return nil
		},
	}

	rootCmd.AddCommand(initCmd, generateCmd, commitCmd)

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func proceseRepo(repoPath string) (changes int, err error) {
	for {
		stepChanges, err := processRepoStep(repoPath)

		if err != nil {
			return changes, nil
		}

		if stepChanges == 0 {
			break
		}

		changes += stepChanges
	}

	return
}

func processRepoStep(repoPath string) (changes int, err error) {
	fsRoot, err := NewFS(repoPath)

	if err != nil {
		fmt.Printf("Error opening %v as git repository: %v\n", repoPath, err)
	}

	err = filepath.Walk(repoPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".go" {
			count, err := processFile(fsRoot, path)

			if err != nil {
				fmt.Printf("Error processing file %v: %v\n", path, err)
				return nil
			}

			changes += count
		}

		return nil
	})

	if err != nil {
		return changes, err
	}

	commitId, err := fsRoot.Commit("TODOs added. I still don't support commit messages.")

	if err != nil {
		return changes, err
	}

	fmt.Printf("Changes committed with commit ID %s\n", commitId)

	err = fsRoot.Push()

	if err != nil {
		fmt.Printf("Error pushing the changes: %v\n", err)
		return changes, err
	}

	return changes, nil
}

func processFile(fsRoot FS, fsPath string) (int, error) {
	// Read the file
	code, err := os.ReadFile(fsPath)
	if err != nil {
		return 0, err
	}

	// Parse the file into an AST
	ast := psi.Parse(fsPath, string(code))

	if ast.Error() != nil {
		return 0, err
	}

	// Process the AST nodes
	updated := codex.ProcessNodes(ast)

	// Convert the AST back to code
	newCode, err := ast.ToCode(updated)
	if err != nil {
		return 0, err
	}

	// Write the new code to a new file
	err = io.WriteFile(fsPath, newCode)
	if err != nil {
		return 0, err
	}

	if newCode != string(code) {
		return 1, nil
	}

	return 0, nil
}

type FS interface {
	fs.FS

	// IsDirty returns true if there are uncommitted changes.
	IsDirty() (bool, error)
	// GetUncommittedChanges returns a string containing the uncommitted changes as a diff.
	GetUncommittedChanges() (string, error)
	// Checkout checks out the given commit.
	Checkout(commit string) error
	// Commit commits the changes with the given message.
	Commit(message string) (commitId string, err error)
	// Push pushes the changes to the remote repository.
	Push() error

	// Path returns the path to the repository.
	Path() string
}

func NewFS(repoPath string) (FS, error) {
	base := os.DirFS(repoPath)

	return &gitFS{
		FS: base,

		path: repoPath,
	}, nil
}

type gitFS struct {
	fs.FS
	path string
}

func (g *gitFS) IsDirty() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.path

	stdout, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(stdout) > 0, nil
}

func (g *gitFS) GetUncommittedChanges() (string, error) {
	cmd := exec.Command("git", "diff")
	cmd.Dir = g.path

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(stdout)), nil
}

func (g *gitFS) Push() error {
	cmd := exec.Command("git", "push")
	cmd.Dir = g.path

	return cmd.Run()
}

func (g *gitFS) Path() string {
	return g.path
}

func (g *gitFS) Checkout(commit string) error {
	cmd := exec.Command("git", "checkout", commit)
	cmd.Dir = g.path

	return cmd.Run()
}

func (g *gitFS) Commit(message string) (commitId string, err error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.path

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}

	if len(stdout) == 0 {
		return "", nil
	}

	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = g.path

	err = cmd.Run()
	if err != nil {
		return "", err
	}

	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = g.path

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, strings.TrimSpace(stderr.String()))
	}

	return strings.Split(out.String(), " ")[1], nil
}
