package codex

import "github.com/greenboxal/agibootstrap/pkg/repofs"

// A Project is the root of a codex project.
// It contains all the information about the project.
// It is also the entry point for all codex operations.
type Project struct {
	rootPath string
	fs       repofs.FS
}

func NewProject(rootPath string) *Project {
	// TODO: Implement
	return &Project{
		rootPath: rootPath,
		fs:       repofs.NewOSFS(rootPath),
	}
}

func (p *Project) RootPath() string { return p.rootPath }

func (p *Project) FS() repofs.FS { return p.fs }

func (p *Project) Sync() error {
	// TODO: Scan all files in the project and update the project's internal state.
	// TODO: It should create psi.FileNode and psi.DirectoryNode objects for each file and directory.
	rootNode, err := psi.NewDirectoryNode(p.fs, "")
	if err != nil {
		return err
	}
	return rootNode.Sync()
}
