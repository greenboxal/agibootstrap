package codex

import (
	"fmt"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
)

// Commit commits the changes made in the project to the file system.
// It checks if there are any changes to commit, prepares the commit message,
// and then performs the commit. If there are no changes, it returns nil.
// It returns an error if any occurs during the commit process.
func (p *Project) Commit() error {
	isDirty, err := p.fs.IsDirty()

	if err != nil {
		return err
	}

	if !isDirty {
		return nil
	}

	diff, err := p.fs.GetStagedChanges()

	if err != nil {
		return err
	}

	commitMessage, err := gpt.PrepareCommitMessage(diff)

	if err != nil {
		return err
	}

	commitId, err := p.fs.Commit(commitMessage)

	if err != nil {
		return err
	}

	fmt.Printf("Changes committed with commit ID %s\n", commitId)

	return nil
}
