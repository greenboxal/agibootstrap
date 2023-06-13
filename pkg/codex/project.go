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
	return &Project{
		rootPath: rootPath,
		fs:       repofs.NewOSFS(rootPath),
	}
}

func (p *Project) RootPath() string { return p.rootPath }

func (p *Project) FS() repofs.FS { return p.fs }

func (p *Project) Sync() error {
	rootNode, err := psi.NewDirectoryNode(p.fs, "")
	if err != nil {
		return err
	}

	err = repofs.Walk(p.fs, "/", func(path string, info repofs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("/", path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			err = psi.CreateDirectoryNode(rootNode, strings.Split(relPath, "/")...)
			if err != nil {
				return err
			}
		} else {
			err = psi.CreateFileNode(rootNode, strings.Split(relPath, "/")...)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
