package codex

import (
	"io/fs"
	"path/filepath"

	"github.com/greenboxal/agibootstrap/pkg/io"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/repofs"
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

	files       []string
	sourceFiles map[string]*psi.SourceFile
}

func NewProject(rootPath string) (*Project, error) {
	root, err := repofs.NewFS(rootPath)

	if err != nil {
		return nil, err
	}

	p := &Project{
		rootPath:    rootPath,
		fs:          root,
		sourceFiles: map[string]*psi.SourceFile{},
	}

	if err := p.Sync(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Project) RootPath() string { return p.rootPath }

func (p *Project) FS() repofs.FS { return p.fs }

func (p *Project) Generate() (changes int, err error) {
	steps := []BuildStep{
		&CodeGenBuildStep{},
		&FixImportsBuildStep{},
		//&FixBuildStep{},
	}

	for {
		stepChanges := 0

		for _, step := range steps {
			processWrapped := func() (result BuildStepResult, err error) {
				defer func() {
					if r := recover(); r != nil {
						if e, ok := r.(error); ok {
							err = e
						} else {
							err = r.(error)
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
	}

	if err = p.fs.Push(); err != nil {
		return
	}

	return
}

func (p *Project) Sync() error {
	p.files = []string{}

	err := filepath.WalkDir(p.rootPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && isGoFile(path) {
			p.files = append(p.files, path)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (p *Project) GetSourceFile(filename string) *psi.SourceFile {
	existing := p.sourceFiles[filename]

	if existing == nil {
		existing = psi.NewSourceFile(filename)
		sourceCode, err := io.ReadFile(filename)

		if err != nil {
			panic(err)
		}

		_, err = existing.Parse(filename, sourceCode)

		if err != nil {
			panic(err)
		}

		p.sourceFiles[filename] = existing
	}

	return existing
}

func isGoFile(path string) bool {
	return filepath.Ext(path) == ".go"
}
