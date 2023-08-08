package codex

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/platform/db/fti"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	tasks "github.com/greenboxal/agibootstrap/pkg/platform/tasks"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
)

const SourceFileEdge psi.TypedEdgeKind[project.SourceFile] = "SourceFile"

type Project struct {
	psi.NodeBase

	logger *zap.SugaredLogger

	uuid   string
	config project.Config

	rootPath string
	rootNode vfs.Node

	LangRegistry     *project.LanguageProvider
	FileTypeRegistry *project.FileTypeProvider
	taskManager      *tasks.Manager
	thoughtRepo      *thoughtdb.Repo
	Fti              *fti.Repository
	analysisManager  *AnalysisManager
	SyncManager      *SyncManager
}

var ProjectType = psi.DefineNodeType[*Project](psi.WithRuntimeOnly())

func NewBareProject() (*Project, error) {
	p := &Project{
		logger: logging.GetLogger("codex"),
	}

	return p, nil
}

// LoadProject creates a new codex project with the given root path.
// It initializes the project file system, repository, and other required data structures.
// It returns a pointer to the created Project object and an error if any.
func LoadProject(ctx context.Context, rootPath string) (*Project, error) {
	p, err := NewBareProject()

	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Project) Init(self psi.Node, projectUuid string) {
	p.uuid = projectUuid

	p.NodeBase.Init(self, psi.WithNodeType(ProjectType))
}

func (p *Project) bootstrap(ctx context.Context) error {
	if err := p.SyncManager.RequestSync(ctx, true); err != nil {
		return err
	}

	for _, lang := range p.config.Project.EnabledLanguages {
		if err := p.bootstrapLanguage(ctx, lang); err != nil {
			return err
		}
	}

	/*for _, mod := range p.config.Modules {
		if err := p.bootstrapModule(ctx, mod); err != nil {
			return err
		}
	}*/

	if err := p.SyncManager.WaitForInitialSync(ctx); err != nil {
		return err
	}

	p.logger.Info("Project bootstrapped")

	return nil
}

func (p *Project) bootstrapLanguage(ctx context.Context, name string) error {
	factory := project.GetLanguageFactory(project.LanguageID(name))

	if factory == nil {
		return errors.Errorf("language %s not found", name)
	}

	p.LangRegistry.Register(factory(p))

	return nil
}

func (p *Project) bootstrapModule(ctx context.Context, mod project.ModuleConfig) error {
	lang := p.LangRegistry.GetLanguage(project.LanguageID(mod.Language))

	if lang == nil {
		return errors.Errorf("language %s not configured", mod.Language)
	}

	/*root, err := psi.Resolve(ctx, p.rootNode, mod.Path)

	if err != nil {
		return errors.Wrapf(err, "failed to resolve module root %s", mod.Path)
	}

	m, err := NewModule(p, mod, lang, root.(*vfs.Directory))

	if err != nil {
		return err
	}

	m.SetParent(p)*/

	return nil
}

func (p *Project) loadConfig(ctx context.Context) error {
	configPath := path.Join(p.rootPath, "Codex.project.toml")

	data, err := os.ReadFile(configPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	if _, err := toml.Decode(string(data), &p.config); err != nil {
		return errors.Wrap(err, "failed to decode config")
	}

	return nil
}

func (p *Project) UUID() string { return p.uuid }

func (p *Project) LanguageProvider() *project.LanguageProvider { return p.LangRegistry }
func (p *Project) FileTypeProvider() *project.FileTypeProvider { return p.FileTypeRegistry }
func (p *Project) TaskManager() *tasks.Manager                 { return p.taskManager }
func (p *Project) LogManager() *thoughtdb.Repo                 { return p.thoughtRepo }
func (p *Project) AnalysisManager() *AnalysisManager           { return p.analysisManager }
func (p *Project) Repo() *fti.Repository                       { return p.Fti }

// RootPath returns the root path of the project.
func (p *Project) RootPath() string   { return p.rootPath }
func (p *Project) RootNode() psi.Node { return p.rootNode }

// Sync synchronizes the project with the file system.
// It scans all files in the project directory and imports
// valid files into the project. Valid files are those that
// have valid file extensions specified in the `validExtensions`
// slice. The function returns an error if any occurs during
// the sync process.
func (p *Project) Sync(ctx context.Context) error {
	return p.SyncManager.RequestSync(ctx, false)
}

// GetSourceFile retrieves the source file with the given filename from the project.
// It returns a pointer to the psi.SourceFile and any error that occurred during the process.
func (p *Project) GetSourceFile(ctx context.Context, filename string) (_ project.SourceFile, err error) {
	relPath, err := filepath.Rel(p.rootPath, filename)

	if err != nil {
		return nil, err
	}

	psiPath := psi.MustParsePath(relPath)
	fileNode, err := psi.ResolvePath(ctx, p.rootNode, psiPath)

	if err != nil {
		return nil, err
	}

	existing := psi.ResolveEdge(fileNode, SourceFileEdge.Singleton())

	if existing == nil {
		if err != nil {
			return nil, err
		}

		lang := p.LangRegistry.ResolveExtension(filepath.Base(filename))

		if lang == nil {
			return nil, fmt.Errorf("failed to resolve language for file %s", filename)
		}

		existing = lang.CreateSourceFile(ctx, filename, &repofs.FsFileHandle{
			FS:   p.vcsFs,
			Path: strings.TrimPrefix(filename, p.rootPath+"/"),
		})

		existing.SetParent(fileNode)
		fileNode.SetEdge(SourceFileEdge.Singleton(), existing)

		err = existing.Load(ctx)

		if err != nil {
			p.logger.Warn(err)
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

func (p *Project) Shutdown(ctx context.Context) error {
	return nil
}
