package repofs

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type FS interface {
	fs.FS

	// IsDirty returns true if there are uncommitted changes.
	IsDirty() (bool, error)
	// GetStagedChanges returns a string containing the staged changes as a diff.
	GetStagedChanges() (string, error)
	// GetUncommittedChanges returns a string containing the uncommitted changes as a diff.
	GetUncommittedChanges() (string, error)
	// Checkout checks out the given commit.
	Checkout(commit string) error
	// Commit commits the changes with the given message.
	Commit(message string) (commitId string, err error)
	// Push pushes the changes to the remote repository.
	Push() error
	// StageAll stages all the changes.
	StageAll() error

	// Path returns the path to the repository.
	Path() string
}

func NewFS(repoPath string) (FS, error) {
	return &gitFS{
		FS: baseFS(repoPath),

		path: repoPath,
	}, nil
}

type baseFS string

func (bfs baseFS) Open(name string) (fs.File, error) {
	fullname, err := bfs.join(name)

	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}

	f, err := os.Open(fullname)

	if err != nil {
		// DirFS takes a string appropriate for GOOS,
		// while the name argument here is always slash separated.
		// bfs.join will have mixed the two; undo that for
		// error reporting.
		err.(*os.PathError).Path = name
		return nil, err
	}

	return f, nil
}

func (bfs baseFS) Stat(name string) (fs.FileInfo, error) {
	fullname, err := bfs.join(name)

	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}

	f, err := os.Stat(fullname)

	if err != nil {
		err.(*fs.PathError).Path = name
		return nil, err
	}

	return f, nil
}

// join returns the path for name in dir.
func (bfs baseFS) join(name string) (string, error) {
	var relPath, absPath string

	if bfs == "" {
		return "", errors.New("repofs: baseFS with empty root")
	}

	if path.IsAbs(name) {
		absPath = name

		rel, relErr := filepath.Rel(string(bfs), name)

		if relErr == nil {
			relPath = rel
		}
	} else {
		relPath = name

		abs, err := filepath.Abs(path.Join(string(bfs), relPath))

		if err == nil {
			absPath = abs
		}
	}

	if relPath == "" || absPath == "" {
		return "", errors.New("file outside of project")
	}

	return absPath, nil
}

type gitFS struct {
	fs.FS
	path string
}

func (g *gitFS) StageAll() error {
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = g.path

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
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

func (g *gitFS) GetStagedChanges() (string, error) {
	cmd := exec.Command("git", "diff", "--cached")
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
