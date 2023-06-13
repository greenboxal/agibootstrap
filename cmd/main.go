package main

import (
	"bytes"
	"fmt"
	fs "io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/greenboxal/agibootstrap/pkg/codex"
	"github.com/greenboxal/agibootstrap/pkg/io"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

func main() {
	// Get a list of all Go files in the current directory and subdirectories
	repoPath := os.Args[1] // Passed as an argument
	fsRoot, err := NewFS(repoPath)

	if err != nil {
		fmt.Printf("Error opening %v as git repository: %v\n", repoPath, err)
	}

	err = filepath.Walk(repoPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".go" {
			processFile(fsRoot, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path %v: %v\n", ".", err)
		return
	}

	commitId, err := fsRoot.Commit("TODOs added. I still don't support commit messages.")

	if err != nil {
		fmt.Printf("Error committing the changes: %v\n", err)
		return
	}

	fmt.Printf("Changes committed with commit ID %s\n", commitId)

	err = fsRoot.Push()

	if err != nil {
		fmt.Printf("Error pushing the changes: %v\n", err)
		return
	}
}

func processFile(fsRoot FS, fsPath string) {
	// Read the file
	code, err := os.ReadFile(fsPath)
	if err != nil {
		fmt.Printf("Error reading the file %s: %v\n", fsPath, err)
		return
	}

	// Parse the file into an AST
	ast := psi.Parse(fsPath, string(code))

	if ast.Error() != nil {
		fmt.Printf("Error parsing the file %s: %v\n", fsPath, ast.Error())
		return
	}

	// Process the AST nodes
	updated := codex.ProcessNodes(ast)

	// Convert the AST back to code
	newCode, err := ast.ToCode(updated)
	if err != nil {
		fmt.Printf("Error converting the AST back to code for file %s: %v\n", fsPath, err)
		return
	}

	// Write the new code to a new file
	err = io.WriteFile(fsPath, newCode)
	if err != nil {
		fmt.Printf("Error writing the new code to a file for %s: %v\n", fsPath, err)
	}
}

type FS interface {
	fs.FS

	Checkout(commit string) error
	Commit(message string) (commitId string, err error)
	Push() error
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
