package repofs

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

type FS interface {
	fs.FS

	Checkout(commit string) error
	Commit(message string) (commitId string, err error)
}

func New(repoPath string) (FS, error) {
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

func (g gitFS) Checkout(commit string) error {
	cmd := exec.Command("git", "checkout", commit)
	cmd.Dir = g.path

	return cmd.Run()
}

func (g gitFS) Commit(message string) (commitId string, err error) {
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
