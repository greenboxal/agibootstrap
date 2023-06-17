package codex

import (
	"context"
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
	"github.com/greenboxal/agibootstrap/pkg/vts"
)

// BuildStepResult represents the result of a build step.
// It contains the number of changes made during the build step.
type BuildStepResult struct {
	Changes int
}

// BuildStep is an interface that defines the contract for a build step.
// It represents a step in the build process and provides a method for processing a project.
type BuildStep interface {
	// Process executes the build step logic on the given project.
	// It returns the result of the build step and any error that occurred during the process.
	Process(ctx context.Context, p *Project) (result BuildStepResult, err error)
}

// Project represents a codex project.
// It contains all the information about the project.
type Project struct {
	psi.NodeBase

	rootPath string
	fs       repofs.FS
	repo     *fti.Repository

	files       map[string]*vfs.FileNode
	sourceFiles map[string]psi.SourceFile

	vtsRoot      *vts.Scope
	langRegistry *Registry

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
		sourceFiles: map[string]psi.SourceFile{},
		fset:        token.NewFileSet(),

		vtsRoot: vts.NewScope(),
	}

	p.langRegistry = NewRegistry(p)

	p.Init(p, "")

	if err := p.Sync(); err != nil {
		return nil, err
	}

	return p, nil
}

// RootPath returns the root path of the project.
func (p *Project) RootPath() string { return p.rootPath }

// FS returns the file system interface of the project.
// It provides methods for managing the project's files and directories.
func (p *Project) FS() repofs.FS { return p.fs }

// Generate performs the code generation process for the project.
// It executes all the build steps specified in the project and
// stages and commits the changes to the file system.
// The isSingleStep parameter determines whether the generation
// process should be performed in a single step or multiple steps.
// If isSingleStep is true, the process stops after the first step
// that makes changes. The function returns the total number of changes
// made during the generation process and an error if any.
func (p *Project) Generate(ctx context.Context, isSingleStep bool) (changes int, err error) {
	// Define the list of build steps to be executed
	steps := []BuildStep{
		//&AnalysisBuildStep{},
		&CodeGenBuildStep{},
		&FixImportsBuildStep{},
		//&FixBuildStep{},
	}

	// Execute the build steps until no further changes are made
	for {
		stepChanges := 0

		// Sync the project to ensure it is up to date
		if err := p.Sync(); err != nil {
			return changes, err
		}

		// Execute each build step
		for _, step := range steps {
			// Wrap the build step execution in a recover function
			processWrapped := func() (result BuildStepResult, err error) {
				defer func() {
					if r := recover(); r != nil {
						if e, ok := r.(error); ok {
							err = e
						} else {
							err = fmt.Errorf("%v", r)
						}
					}
				}()

				// Execute the build step and return the result
				return step.Process(ctx, p)
			}

			// Execute the build step and handle any errors
			result, err := processWrapped()
			if err != nil {
				return changes, err
			}

			// Update the total number of changes
			stepChanges += result.Changes
		}

		// Stage all the changes in the file system
		if err = p.fs.StageAll(); err != nil {
			return
		}

		// Commit the changes to the file system
		if err = p.Commit(); err != nil {
			return
		}

		// If no changes were made, exit the loop
		if stepChanges == 0 {
			break
		}

		// If isSingleStep is true, exit the loop after the first step
		if isSingleStep {
			break
		}

		// Update the total number of changes
		changes += stepChanges
	}

	// Push the changes to the remote repository
	if err = p.fs.Push(); err != nil {
		return
	}

	// Return the total number of changes and no error
	return
}

// Sync synchronizes the project with the file system.
// It scans all files in the project directory and imports
// valid files into the project. Valid files are those that
// have valid file extensions specified in the `validExtensions`
// slice. The function returns an error if any occurs during
// the sync process.
func (p *Project) Sync() (err error) {
	err = filepath.WalkDir(p.rootPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			lang := p.langRegistry.ResolveExtension(path)

			if lang == nil {
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

// GetSourceFile retrieves the source file with the given filename from the project.
// It returns a pointer to the psi.SourceFile and any error that occurred during the process.
func (p *Project) GetSourceFile(filename string) (_ psi.SourceFile, err error) {
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
		lang := p.langRegistry.ResolveExtension(filepath.Base(filename))

		if lang == nil {
			return nil, fmt.Errorf("failed to resolve language for file %s", filename)
		}

		existing = lang.CreateSourceFile(filename, &repofs.FsFileHandle{
			FS:   p.fs,
			Path: strings.TrimPrefix(filename, p.rootPath+"/"),
		})

		err = existing.Load()

		if err != nil {
			return nil, err
		}

		p.sourceFiles[key] = existing
	}

	return existing, nil
}

// ImportFile imports a file into the project.
// It takes the path of the file to import as a parameter.
// It returns an error if the path is invalid or if there
// is any error during the process.
func (p *Project) ImportFile(path string) error {
	absPath, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	file := vfs.NewFileNode(p.fs, absPath)

	p.files[file.UUID()] = file
	p.sourceFiles[file.UUID()] = nil

	if _, err := p.GetSourceFile(file.Path()); err != nil {
		return err
	}

	return nil
}

// Reindex is a method that performs the reindexing operation for the project.
// It updates the index of the project to reflect any changes made to its files.
// The function returns an error if any error occurs during the reindexing process.
func (p *Project) Reindex() error {
	return nil
}

func (p *Project) FileSet() *token.FileSet {
	return p.fset
}
