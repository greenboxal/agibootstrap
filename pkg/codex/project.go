package codex

import (
	"fmt"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/fti"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
	"github.com/greenboxal/agibootstrap/pkg/vfs"
)

type BuildStepResult struct {
	Changes int
}

type BuildStep interface {
	Process(p *Project) (result BuildStepResult, err error)
}

// A Project is the root of a codex project.
// It contains all the information about the project.
// It is also the entry point for all codex operations.
type Project struct {
	rootPath string
	fs       repofs.FS
	repo     *fti.Repository

	files       map[string]*vfs.FileNode
	sourceFiles map[string]*psi.SourceFile

	fset *token.FileSet
}

// NewProject creates a new codex project with the given root path.
// It initializes the project file system, repository, and other required data structures.
// It returns a pointer to the created Project object and an error if any.
func NewProject(rootPath string) (*Project, error) {
	root, err := repofs.NewFS(rootPath)

	if err != nil {
		return nil, err
	}

	repo, err := fti.NewRepository(rootPath)

	if err != nil {
		return nil, err
	}

	p := &Project{
		rootPath:    rootPath,
		fs:          root,
		repo:        repo,
		files:       map[string]*vfs.FileNode{},
		sourceFiles: map[string]*psi.SourceFile{},
		fset:        token.NewFileSet(),
	}

	if err := p.Sync(); err != nil {
		return nil, err
	}

	return p, nil
}

// TODO: Write documentation
func (p *Project) RootPath() string { return p.rootPath }

// TODO: Write documentation
func (p *Project) FS() repofs.FS { return p.fs }

// TODO: Write documentation
func (p *Project) Generate() (changes int, err error) {
	isSingleStep := true

	steps := []BuildStep{
		&CodeGenBuildStep{},
		&FixImportsBuildStep{},
		//&FixBuildStep{},
	}

	for {
		stepChanges := 0

		if err := p.Sync(); err != nil {
			return changes, err
		}

		for _, step := range steps {
			processWrapped := func() (result BuildStepResult, err error) {
				defer func() {
					if r := recover(); r != nil {
						if e, ok := r.(error); ok {
							err = e
						} else {
							err = fmt.Errorf("%v", e)
						}
					}
				}()

				return step.Process(p)
			}

			result, err := processWrapped()

			if err != nil {
				return changes, err
			}

			stepChanges += result.Changes
		}

		if stepChanges == 0 {
			break
		}

		changes += stepChanges

		if err = p.fs.StageAll(); err != nil {
			return
		}

		if err = p.Commit(); err != nil {
			return
		}

		if isSingleStep {
			break
		}
	}

	if err = p.fs.Push(); err != nil {
		return
	}

	return
}

var validExtensions = []string{".go"}

func (p *Project) Sync() (err error) {

	err = filepath.WalkDir(p.rootPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			valid := false

			for _, ext := range validExtensions {
				if strings.HasSuffix(path, ext) {
					valid = true
					break
				}
			}

			if !valid {
				return nil
			}

			err := p.ImportFile(path)

			if err != nil {
				return err
			}
		}

		return nil
	})

	return nil
}

func (p *Project) GetSourceFile(filename string) (_ *psi.SourceFile, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = r.(error)
			}

			err = errors.Wrap(err, "failed to get source file "+filename)
		}
	}()

	absPath, err := filepath.Abs(filename)

	if err != nil {
		return nil, err
	}

	key := strings.ToLower(absPath)

	existing := p.sourceFiles[key]

	if existing == nil {
		existing = psi.NewSourceFile(p.fset, filename, repofs.FsFileHandle{
			FS:   p.fs,
			Path: strings.TrimPrefix(filename, p.rootPath+"/"),
		})

		err = existing.Load()

		if err != nil {
			return nil, err
		}

		p.sourceFiles[filename] = existing
	}

	return existing, nil
}

func (p *Project) ImportFile(path string) error {
	absPath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	file := vfs.NewFileNode(absPath)

	p.files[file.Key] = file
	p.sourceFiles[file.Key] = nil

	if _, err := p.GetSourceFile(file.Path); err != nil {
		return err
	}

	return nil
}

func (p *Project) Reindex() error {
	return nil
}
