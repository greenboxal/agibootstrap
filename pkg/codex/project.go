package codex

import (
	"context"
	"fmt"
	"go/token"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/greenboxal/agibootstrap/pkg/indexing"

	"github.com/greenboxal/agibootstrap/pkg/codex/vts"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/fti"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	tasks2 "github.com/greenboxal/agibootstrap/pkg/platform/tasks"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

const SourceFileEdge psi.TypedEdgeKind[psi.SourceFile] = "SourceFile"

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

	g *indexing.IndexedGraph

	fs         repofs.FS
	fsRootNode *vfs.DirectoryNode

	repo *fti.Repository
	tm   *tasks2.Manager

	rootPath string
	rootNode *vfs.DirectoryNode

	vts          *vts.Scope
	langRegistry *project.Registry

	fset *token.FileSet

	currentSyncTaskMutex sync.Mutex
	currentSyncTask      tasks2.Task
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
		rootPath: rootPath,

		fs:   root,
		repo: repo,

		fset: token.NewFileSet(),
		vts:  vts.NewScope(),
	}

	p.Init(p, "")

	p.langRegistry = project.NewRegistry(p)

	p.rootNode = vfs.NewDirectoryNode(p.fs, p.rootPath, "srcs")
	p.rootNode.SetParent(p)

	p.tm = tasks2.NewManager()
	p.tm.PsiNode().SetParent(p)

	p.g = indexing.NewIndexedGraph(p)
	p.g.Add(p)

	if err := p.Sync(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Project) TaskManager() *tasks2.Manager { return p.tm }

// RootPath returns the root path of the project.
func (p *Project) RootPath() string   { return p.rootPath }
func (p *Project) RootNode() psi.Node { return p.rootNode }

// FS returns the file system interface of the project.
// It provides methods for managing the project's files and directories.
func (p *Project) FS() repofs.FS { return p.fs }

func (p *Project) Graph() psi.Graph                    { return p.g }
func (p *Project) LanguageProvider() *project.Registry { return p.langRegistry }

func (p *Project) Repo() *fti.Repository { return p.repo }

func (p *Project) FileSet() *token.FileSet { return p.fset }

// Sync synchronizes the project with the file system.
// It scans all files in the project directory and imports
// valid files into the project. Valid files are those that
// have valid file extensions specified in the `validExtensions`
// slice. The function returns an error if any occurs during
// the sync process.
func (p *Project) Sync() error {
	p.currentSyncTaskMutex.Lock()
	defer p.currentSyncTaskMutex.Unlock()

	if p.currentSyncTask != nil {
		return nil
	}

	task := p.tm.SpawnTask(context.Background(), func(progress tasks2.TaskProgress) error {
		maxDepth := 0
		count := 0

		defer progress.Update(maxDepth*maxDepth+count+1, maxDepth*maxDepth+count+1)

		return psi.Walk(p.rootNode, func(cursor psi.Cursor, entering bool) error {
			if cursor.Depth() > maxDepth {
				maxDepth = cursor.Depth()
			}

			progress.Update(cursor.Depth()*cursor.Depth()+count, maxDepth*maxDepth+count+1)

			n := cursor.Node()

			time.Sleep(time.Millisecond * 100)

			if n, ok := n.(*vfs.DirectoryNode); ok && entering {
				count++

				cursor.WalkChildren()

				return n.Sync(func(path string) bool {
					return !p.repo.IsIgnored(path)
				})
			} else {
				cursor.SkipChildren()
			}

			return nil
		})
	})

	p.currentSyncTask = task

	go func() {
		<-task.Done()

		p.currentSyncTaskMutex.Lock()
		defer p.currentSyncTaskMutex.Unlock()

		if p.currentSyncTask == task {
			p.currentSyncTask = nil
		}
	}()

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

	relPath, err := filepath.Rel(p.rootPath, filename)

	if err != nil {
		return nil, err
	}

	psiPath := psi.MustParsePath(relPath)
	fileNode, err := psi.ResolvePath(p.rootNode, psiPath)

	if err != nil {
		return nil, err
	}

	existing := psi.ResolveEdge(fileNode, SourceFileEdge.Singleton())

	if existing == nil {
		if err != nil {
			return nil, err
		}

		lang := p.langRegistry.ResolveExtension(filepath.Base(filename))

		if lang == nil {
			return nil, fmt.Errorf("failed to resolve language for file %s", filename)
		}

		existing = lang.CreateSourceFile(filename, &repofs.FsFileHandle{
			FS:   p.fs,
			Path: strings.TrimPrefix(filename, p.rootPath+"/"),
		})

		existing.SetParent(fileNode)
		fileNode.SetEdge(SourceFileEdge.Singleton(), existing)

		err = existing.Load()

		if err != nil {
			return nil, err
		}
	}

	return existing, nil
}

// Reindex is a method that performs the reindexing operation for the project.
// It updates the index of the project to reflect any changes made to its files.
// The function returns an error if any error occurs during the reindexing process.
func (p *Project) Reindex() error {
	return nil
}
